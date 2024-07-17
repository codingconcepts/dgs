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

	generatedMu sync.RWMutex
	generated   map[string]int
}

// NewDataGenerator returns a pointer to a new instance of DataGenerator.
func NewDataGenerator(db *pgxpool.Pool, logger zerolog.Logger, config model.Config, workers, batch int) *DataGenerator {
	return &DataGenerator{
		db:        db,
		logger:    logger,
		config:    config,
		workers:   workers,
		batch:     batch,
		generated: map[string]int{},
	}
}

func (g *DataGenerator) calculateIterations() map[string]iteration {
	iterations := map[string]iteration{}

	for _, t := range g.config.Tables {
		i := calculateIteration(t, g.batch, g.workers)
		g.logger.Info().
			Str("table", t.Name).
			Int("rows", t.Rows).
			Int("batch", i.batch).
			Int("times", i.times).
			Msg("iteration")

		iterations[t.Name] = i
	}

	return iterations
}

func calculateIteration(table model.Table, batch, workers int) iteration {
	if table.Rows < batch/workers {
		return iteration{
			times: 1,
			batch: table.Rows / workers,
		}
	}

	if table.Rows <= batch {
		return iteration{
			batch: table.Rows / workers,
			times: 1,
		}
	}

	if table.Rows/batch/workers == 0 {
		return iteration{
			batch: batch,
			times: 1,
		}
	}

	return iteration{
		batch: batch,
		times: table.Rows / batch / workers,
	}
}

type iteration struct {
	times int
	batch int
}

func (g *DataGenerator) Generate() error {
	iterations := g.calculateIterations()

	g.logger.Info().
		Int("workers", g.workers).
		Int("batch", g.batch).
		Msg("generating")

	var eg errgroup.Group

	for w := 0; w < g.workers; w++ {
		workerID := w + 1
		eg.Go(func() error {
			if err := g.generateWorker(iterations, workerID); err != nil {
				return fmt.Errorf("generate worker: %w", err)
			}

			return nil
		})
	}

	if err := eg.Wait(); err != nil {
		return fmt.Errorf("generating data: %w", err)
	}

	for k, v := range g.generated {
		g.logger.Info().
			Str("table", k).
			Int("rows", v).
			Msg("finished generating")
	}

	return nil
}

func (g *DataGenerator) generateWorker(iterations map[string]iteration, wid int) error {
	g.logger.Info().Int("worker id", wid).Msg("scheduled")

	db, err := g.db.Acquire(context.Background())
	if err != nil {
		return fmt.Errorf("acquiring database connection: %w", err)
	}
	defer db.Release()

	g.logger.Info().Int("worker id", wid).Msg("started")

	data := model.NewIterationData()
	for _, table := range g.config.Tables {
		iter := iterations[table.Name]

		for i := 0; i < iter.times; i++ {
			// Generate rows.
			rows, err := g.generateRows(table, data, g.batch)
			if err != nil {
				return fmt.Errorf("generating rows: %w", err)
			}

			// Write rows.
			if err = g.writeRows(db, table, data, rows); err != nil {
				return fmt.Errorf("writing rows: %w", err)
			}

			g.generatedMu.Lock()
			g.generated[table.Name] += g.batch
			g.logger.Info().
				Str("table", table.Name).
				Int("worker id", wid).
				Str("generated", humanize.Comma(int64(iter.batch))).
				Str("total", humanize.Comma(int64(g.generated[table.Name]))).
				Msg("progress")
			g.generatedMu.Unlock()
		}
	}

	g.logger.Info().Int("worker id", wid).Msg("finished")

	return nil
}

func (g *DataGenerator) generateRows(table model.Table, data *model.IterationData, batch int) ([][]any, error) {
	rows := [][]any{}

	for i := 0; i < batch; i++ {
		row, err := g.generateRow(table.Columns, data)
		if err != nil {
			return nil, fmt.Errorf("generating row: %w", err)
		}
		rows = append(rows, row)
	}

	return rows, nil
}

func (g *DataGenerator) generateRow(columns []model.Column, data *model.IterationData) ([]any, error) {
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
			row = append(row, data.GetValue(c.Ref))

		case model.ColumnTypeInc:
			row = append(row, c.NextID())

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

func (g *DataGenerator) writeRows(db *pgxpool.Conn, table model.Table, data *model.IterationData, rows [][]any) error {
	stmt, err := query.BuildInsert(table, rows)
	if err != nil {
		return fmt.Errorf("building insert: %w", err)
	}
	g.logger.Debug().Str("stmt", stmt).Msg("running insert")

	timeout, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	if _, err := db.Exec(timeout, stmt, lo.Flatten(rows)...); err != nil {
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

		data.AddData(rows, table.Name, column, index)
		g.logger.Debug().
			Str("column", column).
			Any("values", data.GetValues(fmt.Sprintf("%s.%s", table.Name, column))).
			Msg("persisting ref column")
	}

	return nil
}
