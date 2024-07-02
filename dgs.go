package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
	"github.com/samber/lo"
	"golang.org/x/sync/errgroup"

	"github.com/codingconcepts/dgs/pkg/model"
	"github.com/codingconcepts/dgs/pkg/query"
	"github.com/codingconcepts/dgs/pkg/random"
)

var (
	logger zerolog.Logger
)

func main() {
	url := flag.String("url", "", "database connection string")
	config := flag.String("config", "", "absolute or relative path to the config file")
	debug := flag.Bool("debug", false, "enable debug logging")
	batch := flag.Int("batch", 10000, "query and insert batch size")
	workers := flag.Int("workers", 4, "number of workers to run concurrently")
	flag.Parse()

	if *config == "" || *url == "" {
		flag.Usage()
		os.Exit(2)
	}

	logger = zerolog.New(zerolog.ConsoleWriter{
		Out: os.Stderr,
		PartsExclude: []string{
			zerolog.TimestampFieldName,
		},
	}).Level(lo.Ternary(*debug, zerolog.DebugLevel, zerolog.InfoLevel))

	logger.Debug().Msgf("reading file: %s", *config)
	file, err := os.ReadFile(*config)
	if err != nil {
		logger.Fatal().Msgf("error reading config file: %v", err)
	}

	logger.Debug().Msgf("parsing file: %s", *config)
	c, err := model.ParseConfig(string(file), logger)
	if err != nil {
		logger.Fatal().Msgf("error parsing config file: %v", err)
	}

	db, err := pgxpool.New(context.Background(), *url)
	if err != nil {
		logger.Fatal().Msgf("error connecting to database: %v", err)
	}
	defer db.Close()

	logger.Debug().Msg("generating data")
	if err = generate(db, c, *workers, *batch); err != nil {
		logger.Fatal().Msgf("error generating data: %v", err)
	}

	logger.Info().Msg("done")
}

func generate(db *pgxpool.Pool, config model.Config, workers, batch int) error {
	for _, table := range config.Tables {
		logger.Info().Str("table", table.Name).Msg("generating table")

		refs, err := populateRefs(db, workers, table.Columns, batch)
		if err != nil {
			return fmt.Errorf("populating refs: %w", err)
		}

		var progressMu sync.Mutex
		totalProcessed := 0

		eg := new(errgroup.Group)
		rowGroups := splitRows(table.Rows, workers)

		for _, rowsCount := range rowGroups {
			rowsCount := rowsCount // capture range variable
			eg.Go(func() error {
				var rows [][]any
				for i := 0; i < rowsCount; i++ {
					row, err := generateRow(table.Columns, refs)
					if err != nil {
						return fmt.Errorf("generating row: %w", err)
					}
					rows = append(rows, row)

					// Flush if batch size reached.
					if len(rows) == batch {
						progressMu.Lock()
						totalProcessed += len(rows)
						logger.Info().Msgf("writing rows %d/%d", totalProcessed, table.Rows)
						progressMu.Unlock()

						if err := writeRows(db, table, rows); err != nil {
							return fmt.Errorf("writing rows: %w", err)
						}
						rows = nil
					}
				}

				// Flush any stragglers.
				if len(rows) > 0 {
					progressMu.Lock()
					totalProcessed += len(rows)
					logger.Info().Msgf("writing rows %d/%d", totalProcessed, table.Rows)
					progressMu.Unlock()

					if err := writeRows(db, table, rows); err != nil {
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
func splitRows(totalRows, workerCount int) []int {
	rowsPerWorker := totalRows / workerCount
	rowGroups := make([]int, workerCount)
	for i := 0; i < workerCount; i++ {
		rowGroups[i] = rowsPerWorker
	}

	// Distribute the remainder.
	for i := 0; i < totalRows%workerCount; i++ {
		rowGroups[i]++
	}
	return rowGroups
}

func populateRefs(db *pgxpool.Pool, workers int, columns []model.Column, batch int) (map[string][]any, error) {
	var resultsMu sync.Mutex
	results := map[string][]any{}

	eg := errgroup.Group{}
	eg.SetLimit(workers)

	for _, c := range columns {
		if c.Ref == "" {
			continue
		}

		c := c
		eg.Go(func() error {
			logger.Debug().Str("column", c.Name).Msg("populating ref")
			values, err := populateRef(db, c, batch)
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

func populateRef(db *pgxpool.Pool, column model.Column, batch int) ([]any, error) {
	const stmtFmt = `SELECT %s
									 FROM %s
									 ORDER BY RANDOM()
									 LIMIT $1`

	parts := strings.Split(column.Ref, ".")
	stmt := fmt.Sprintf(stmtFmt, parts[1], parts[0])

	rows, err := db.Query(context.Background(), stmt, batch)
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

func generateRow(columns []model.Column, refs map[string][]any) ([]any, error) {
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

	logger.Debug().Msgf("row: %+v", row)
	return row, nil
}

func generateRange(c model.Column) (any, error) {
	switch x := strings.ToLower(c.Range); x {
	case "int":
		return random.Int(c.Min, c.Max)

	case "float":
		return random.Float(c.Min, c.Max)

	case "bytes":
		return random.Bytes(c.Min, c.Max)

	case "timestamp":
		v, err := random.Timestamp(c.Min, c.Max)
		if err != nil {
			return nil, fmt.Errorf("generating timestamp: %w", err)
		}
		return formatValue(c.Format, v), nil

	default:
		return nil, fmt.Errorf("invalid type for range: %q", x)
	}
}

func generateValue(c model.Column) (any, error) {
	value := c.Value

	// Look for quick single-replacements.
	if v, ok := random.Replacements[value]; ok {
		return formatValue(c.Format, v()), nil
	}

	// Process multiple-replacements.
	for k, v := range random.Replacements {
		if strings.Contains(value, k) {
			value = strings.ReplaceAll(value, k, formatValue(c.Format, v()))
		}
	}

	return value, nil
}

func formatValue(format string, value any) string {
	// Check if the value implements the formatter interface and use that first,
	// otherwise, just perform a simple string format.
	if format != "" {
		if f, ok := value.(model.Formatter); ok {
			return f.Format(format)
		}

		return fmt.Sprintf(format, value)
	}

	return fmt.Sprintf("%v", value)
}

func writeRows(db *pgxpool.Pool, table model.Table, rows [][]any) error {
	stmt, err := query.BuildInsert(table, rows)
	if err != nil {
		return fmt.Errorf("building insert: %w", err)
	}
	logger.Debug().Str("stmt", stmt).Msg("running insert")

	timeout, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	if _, err := db.Exec(timeout, stmt, lo.Flatten(rows)...); err != nil {
		return fmt.Errorf("executing query: %w", err)
	}

	return nil
}
