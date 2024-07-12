package commands

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/codingconcepts/dgs/pkg/model"
	"github.com/codingconcepts/dgs/pkg/query"
	"github.com/codingconcepts/dgs/pkg/random"
	"github.com/codingconcepts/dgs/pkg/ui"
	"github.com/dustin/go-humanize"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
	"github.com/samber/lo"
	"golang.org/x/sync/errgroup"
)

// DataGenerator holds the runtime dependencies of gen data.
type DataGenerator struct {
	db      *pgxpool.Pool
	logger  zerolog.Logger
	config  model.Config
	workers int
	batch   int
}

// NewDataGenerator returns a pointer to a new instance of DataGenerator.
func NewDataGenerator(db *pgxpool.Pool, logger zerolog.Logger, config model.Config, workers, batch int) *DataGenerator {
	return &DataGenerator{
		db:      db,
		logger:  logger,
		config:  config,
		workers: workers,
		batch:   batch,
	}
}

// Generate data.
func (g *DataGenerator) Generate() error {
	for _, table := range g.config.Tables {
		g.logger.Info().Str("table", table.Name).Msg("generating table")

		refs, err := g.populateRefs(table.Columns)
		if err != nil {
			return fmt.Errorf("populating refs: %w", err)
		}

		var progressMu sync.Mutex
		var totalProcessed int

		eg := new(errgroup.Group)
		rowGroups := g.splitRows(table.Rows)

		for _, rowsCount := range rowGroups {
			rowsCount := rowsCount
			eg.Go(func() error {
				var rows [][]any
				for i := 0; i < rowsCount; i++ {
					row, err := g.generateRow(table.Columns, refs)
					if err != nil {
						return fmt.Errorf("generating row: %w", err)
					}
					rows = append(rows, row)

					// Flush if batch size reached.
					if len(rows) == g.batch {
						progressMu.Lock()
						totalProcessed += len(rows)
						g.logger.Info().Msgf("writing rows %s/%s", humanize.Comma(int64(totalProcessed)), humanize.Comma(int64(table.Rows)))
						progressMu.Unlock()

						if err := g.writeRows(table, rows); err != nil {
							return fmt.Errorf("writing rows: %w", err)
						}
						rows = nil

						// Renew refs.
						refs, err = g.populateRefs(table.Columns)
						if err != nil {
							return fmt.Errorf("populating refs: %w", err)
						}
					}
				}

				// Flush any stragglers.
				if len(rows) > 0 {
					progressMu.Lock()
					totalProcessed += len(rows)
					g.logger.Info().Msgf("writing rows %d/%d", totalProcessed, table.Rows)
					progressMu.Unlock()

					if err := g.writeRows(table, rows); err != nil {
						return fmt.Errorf("writing rows: %w", err)
					}
				}
				return nil
			})
		}

		if err := eg.Wait(); err != nil {
			return fmt.Errorf("error generating data: %w", err)
		}
	}

	return nil
}

// splitRows distributes the total rows to generate for a table across the
// workers participating in the load.
func (g *DataGenerator) splitRows(totalRows int) []int {
	rowsPerWorker := totalRows / g.workers
	rowGroups := make([]int, g.workers)
	for i := 0; i < g.workers; i++ {
		rowGroups[i] = rowsPerWorker
	}

	// Distribute the remainder.
	for i := 0; i < totalRows%g.workers; i++ {
		rowGroups[i]++
	}
	return rowGroups
}

func (g *DataGenerator) populateRefs(columns []model.Column) (map[string][]any, error) {
	timer := ui.NewTimer("populateRefs", g.logger)
	defer timer.Log()

	var resultsMu sync.Mutex
	results := map[string][]any{}

	eg := errgroup.Group{}
	eg.SetLimit(g.workers)

	for _, c := range columns {
		if c.Ref == "" {
			continue
		}

		c := c
		eg.Go(func() error {
			g.logger.Debug().Str("column", c.Name).Msg("populating ref")
			values, err := g.populateRef(c)
			if err != nil {
				return fmt.Errorf("populating ref for column %q", c.Name)
			}

			resultsMu.Lock()
			results[c.Name] = values
			resultsMu.Unlock()

			return nil
		})
	}

	return results, eg.Wait()
}

func (g *DataGenerator) populateRef(column model.Column) ([]any, error) {
	const stmtFmt = `SELECT %s
									 FROM %s
									 ORDER BY RANDOM()
									 LIMIT $1`

	parts := strings.Split(column.Ref, ".")
	stmt := fmt.Sprintf(stmtFmt, parts[1], parts[0])

	rows, err := g.db.Query(context.Background(), stmt, g.batch)
	if err != nil {
		return nil, fmt.Errorf("querying ref table: %w", err)
	}

	var results []any
	var r any

	for rows.Next() {
		if err = rows.Scan(&r); err != nil {
			return nil, fmt.Errorf("scanning ref row: %w", err)
		}
		results = append(results, r)
	}

	return results, nil
}

func (g *DataGenerator) generateRow(columns []model.Column, refs map[string][]any) ([]any, error) {
	row := []any{}

	for _, c := range columns {
		switch c.Mode {
		case model.ColumnTypeValue:
			val, err := generateValue(c)
			if err != nil {
				return nil, fmt.Errorf("generating value: %w", err)
			}
			row = append(row, val)

		case model.ColumnTypeRange:
			val, err := generateRange(c)
			if err != nil {
				return nil, fmt.Errorf("generating range: %w", err)
			}
			row = append(row, val)

		case model.ColumnTypeSet:
			row = append(row, lo.Sample(c.Set))

		case model.ColumnTypeRef:
			row = append(row, lo.Sample(refs[c.Name]))

		default:
			return nil, fmt.Errorf("invalid column mode: %q", c.Mode)
		}
	}

	g.logger.Debug().Msgf("row: %+v", row)
	return row, nil
}

func generateRange(c model.Column) (any, error) {
	switch x := strings.ToLower(c.Range); x {
	case "int":
		var x model.IntRange
		if err := c.Props.Unmarshal(&x); err != nil {
			return nil, fmt.Errorf("decoding int range props: %w", err)
		}
		return random.Int(x.Min, x.Max), nil

	case "float":
		var x model.FloatRange
		if err := c.Props.Unmarshal(&x); err != nil {
			return nil, fmt.Errorf("decoding float range props: %w", err)
		}
		return random.Float(x.Min, x.Max), nil

	case "bytes":
		var x model.ByteRange
		if err := c.Props.Unmarshal(&x); err != nil {
			return nil, fmt.Errorf("decoding bytes range props: %w", err)
		}
		return random.Bytes(x.Min, x.Max)

	case "timestamp":
		var x model.TimestampRange
		if err := c.Props.Unmarshal(&x); err != nil {
			return nil, fmt.Errorf("decoding timestamp range props: %w", err)
		}

		v := random.Timestamp(x.Min, x.Max)
		return v.Format(x.Format), nil

	case "point":
		var x model.PointRange
		if err := c.Props.Unmarshal(&x); err != nil {
			return nil, fmt.Errorf("decoding point range props: %w", err)
		}

		lon, lat := random.Point(x.Lat, x.Lon, float64(x.DistanceKM))
		return model.Point{Lat: lat, Lon: lon}, nil

	default:
		return nil, fmt.Errorf("invalid type for range: %q", x)
	}
}

func generateValue(c model.Column) (any, error) {
	value := c.Value

	// Look for quick single-replacements.
	if v, ok := random.Replacements[value]; ok {
		return v(), nil
	}

	// Process multiple-replacements.
	for k, v := range random.Replacements {
		if strings.Contains(value, k) {
			value = strings.ReplaceAll(value, k, fmt.Sprintf("%v", v()))
		}
	}

	return value, nil
}

func (g *DataGenerator) writeRows(table model.Table, rows [][]any) error {
	stmt, err := query.BuildInsert(table, rows)
	if err != nil {
		return fmt.Errorf("building insert: %w", err)
	}
	g.logger.Debug().Str("stmt", stmt).Msg("running insert")

	timeout, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	if _, err := g.db.Exec(timeout, stmt, lo.Flatten(rows)...); err != nil {
		return fmt.Errorf("executing query: %w", err)
	}

	return nil
}
