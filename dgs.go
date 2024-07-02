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
	if err = generate(db, c); err != nil {
		logger.Fatal().Msgf("error generating data: %v", err)
	}

	logger.Info().Msg("done")
}

func generate(db *pgxpool.Pool, config model.Config) error {
	for _, table := range config.Tables {
		logger.Info().Str("table", table.Name).Msg("generating table")

		var rows [][]any

		refs, err := populateRefs(db, config.Workers, table.Columns, config.Batch)
		if err != nil {
			return fmt.Errorf("populating refs: %w", err)
		}

		for i := 0; i < table.Rows; i++ {
			row, err := generateRow(table.Columns, refs)
			if err != nil {
				return fmt.Errorf("generating row: %w", err)
			}

			rows = append(rows, row)

			// Flush if batch size reached.
			if len(rows) == config.Batch {
				logger.Info().Msgf("writing rows %d/%d", i, table.Rows)
				if err := writeRows(db, table, rows); err != nil {
					return fmt.Errorf("writing rows: %w", err)
				}
				rows = nil
			}
		}

		// Flush any stragglers.
		if len(rows) > 0 {
			logger.Info().Msg("writing remaining rows")
			if err := writeRows(db, table, rows); err != nil {
				return fmt.Errorf("writing rows: %w", err)
			}
			rows = nil
		}
	}

	return nil
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

	// Process multipe-replacements.
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
