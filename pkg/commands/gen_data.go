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
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
	"github.com/samber/lo"
)

// DataGenerator holds the runtime dependencies of gen data.
type DataGenerator struct {
	db      *pgxpool.Pool
	logger  zerolog.Logger
	config  model.Config
	workers int
	batch   int

	generatedMu   sync.RWMutex
	generated     map[string]int
	iterationData *model.IterationData
}

// NewDataGenerator returns a pointer to a new instance of DataGenerator.
func NewDataGenerator(db *pgxpool.Pool, logger zerolog.Logger, config model.Config, workers, batch int) *DataGenerator {
	return &DataGenerator{
		db:            db,
		logger:        logger,
		config:        config,
		workers:       workers,
		batch:         batch,
		iterationData: model.NewIterationData(),
		generated:     map[string]int{},
	}
}

func calculateIterations(tables []model.Table, batchSize int) (map[string]int, error) {
	minRows := lo.MinBy(tables, func(a, b model.Table) bool {
		return a.Rows < b.Rows
	})

	if minRows.Rows < batchSize {
		return nil, fmt.Errorf("smallest table must have at least %d rows", batchSize)
	}

	return lo.SliceToMap(tables, func(t model.Table) (string, int) {
		return t.Name, t.Rows / batchSize / (minRows.Rows / batchSize)
	}), nil
}

func (g *DataGenerator) Generate() error {
	iterations, err := calculateIterations(g.config.Tables, g.batch)
	if err != nil {
		return fmt.Errorf("calculating iteration counts: %w", err)
	}

	for {
		for _, t := range g.config.Tables {
			g.logger.Info().Str("table", t.Name).Msg("generating table")

			for i := 0; i < iterations[t.Name]; i++ {
				// Generate rows.
				rows, err := g.generateRows(t)
				if err != nil {
					return fmt.Errorf("generating rows: %w", err)
				}

				// Write rows.
				if err = g.writeRows(t, rows); err != nil {
					return fmt.Errorf("writing rows: %w", err)
				}

				g.generatedMu.Lock()
				g.generated[t.Name] += g.batch
				g.generatedMu.Unlock()
			}
		}

		if g.finished() {
			break
		}
	}

	return nil
}

func (g *DataGenerator) finished() bool {
	g.generatedMu.RLock()
	defer g.generatedMu.RLock()

	for _, t := range g.config.Tables {
		g, ok := g.generated[t.Name]
		if !ok {
			return false
		}

		if g < t.Rows {
			return false
		}
	}

	return true
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

func (g *DataGenerator) generateRows(table model.Table) ([][]any, error) {
	rows := [][]any{}

	for i := 0; i < g.batch; i++ {
		row, err := g.generateRow(table.Columns)
		if err != nil {
			return nil, fmt.Errorf("generating row: %w", err)
		}
		rows = append(rows, row)
	}

	return rows, nil
}

func (g *DataGenerator) generateRow(columns []model.Column) ([]any, error) {
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
			row = append(row, g.iterationData.GetValue(c.Ref))

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

	// Return the generated rows that match the columns that other tables reference.
	for _, column := range table.RefColumns {
		_, index, ok := lo.FindIndexOf(table.Columns, func(c model.Column) bool {
			return c.Name == column
		})

		if !ok {
			continue
		}

		g.iterationData.AddData(rows, table.Name, column, index)
		g.logger.Debug().
			Str("column", column).
			Any("values", g.iterationData.GetValues(fmt.Sprintf("%s.%s", table.Name, column))).
			Msg("persisting ref column")
	}

	return nil
}
