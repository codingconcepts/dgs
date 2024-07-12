package commands

import (
	"context"
	_ "embed"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/codingconcepts/dgs/pkg/model"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/samber/lo"
)

//go:embed column_definitions.sql
var columnDefinitionsStmt string

// GenerateConfig creates a config object
func GenerateConfig(db *pgxpool.Pool, schema string, rowCounts []string) (model.Config, error) {
	rowCountMap, err := parseRowCounts(rowCounts)
	if err != nil {
		return model.Config{}, fmt.Errorf("parsing row count: %w", err)
	}

	columns, err := fetchColumnDefinitions(db, schema)
	if err != nil {
		return model.Config{}, fmt.Errorf("fetching column definitions: %w", err)
	}

	tables, err := toConfigs(columns, rowCountMap)
	if err != nil {
		return model.Config{}, fmt.Errorf("converting column defintions to config: %w", err)
	}

	return model.Config{
		Tables: tables,
	}, nil
}

func parseRowCounts(rowCounts []string) (map[string]int, error) {
	m := map[string]int{}
	for _, rc := range rowCounts {
		parts := strings.Split(rc, ":")
		if len(parts) != 2 {
			continue
		}

		tableName := parts[0]

		rowCount, err := strconv.Atoi(parts[1])
		if err != nil {
			return nil, fmt.Errorf("parsing count %q: %w", rc, err)
		}

		m[tableName] = rowCount
	}

	return m, nil
}

type columnDefinition struct {
	TableName        string
	ColumnName       string
	Default          *string
	Nullable         string
	DataType         string
	UserDefintedType *[]string
	ForeignKey       *string
}

func fetchColumnDefinitions(db *pgxpool.Pool, schema string) ([]columnDefinition, error) {
	rows, err := db.Query(context.Background(), columnDefinitionsStmt, schema)
	if err != nil {
		return nil, fmt.Errorf("querying column definitions: %w", err)
	}

	var definitions []columnDefinition
	var d columnDefinition

	for rows.Next() {
		if err = rows.Scan(&d.TableName, &d.ColumnName, &d.Default, &d.Nullable, &d.DataType, &d.UserDefintedType, &d.ForeignKey); err != nil {
			return nil, fmt.Errorf("scanning column definition: %w", err)
		}
		definitions = append(definitions, d)
	}

	return definitions, nil
}

func toConfigs(definitions []columnDefinition, rowCountMap map[string]int) ([]model.Table, error) {
	groups := lo.GroupBy(definitions, func(d columnDefinition) string {
		return d.TableName
	})

	var tables []model.Table
	for _, t := range groups {
		table := model.Table{
			Name: t[0].TableName,
			Rows: 100000,
		}

		// Apply row count if provided.
		if rowCount, ok := rowCountMap[table.Name]; ok {
			table.Rows = rowCount
		}

		for _, c := range t {
			if c.ForeignKey != nil {
				table.Columns = append(table.Columns, createRefColumn(c))
				continue
			}

			if c.DataType == "enum" {
				if c.UserDefintedType == nil {
					return nil, fmt.Errorf("missing values for enum column %q", c.ColumnName)
				}
				table.Columns = append(table.Columns, createEnumColumn(c))
				continue
			}

			column, err := createRegularColumn(c)
			if err != nil {
				return nil, fmt.Errorf("creating column: %w", err)
			}
			table.Columns = append(table.Columns, column)
		}

		tables = append(tables, table)
	}

	tables, err := model.SortTables(tables)
	if err != nil {
		return nil, fmt.Errorf("sorting tables into relational order: %w", err)
	}

	return tables, nil
}

func createRefColumn(c columnDefinition) model.Column {
	return model.Column{
		Name: c.ColumnName,
		Mode: model.ColumnTypeRef,
		Ref:  *c.ForeignKey,
	}
}

func createEnumColumn(c columnDefinition) model.Column {
	return model.Column{
		Name: c.ColumnName,
		Mode: model.ColumnTypeSet,
		Set:  *c.UserDefintedType,
	}
}

func createRegularColumn(c columnDefinition) (model.Column, error) {
	var err error

	column := model.Column{
		Name: c.ColumnName,
	}

	switch c.DataType {
	case "uuid":
		column.Value = "${uuid}"

	case "text":
		column.Value = valueForTextColumn(c)

	case "int8":
		column.Range = "int"
		if column.Props, err = model.NewRawMessage(model.IntRange{Min: 1, Max: 1000000}); err != nil {
			return model.Column{}, fmt.Errorf("creating props for int range: %w", err)
		}

	case "numeric":
		column.Range = "float"
		if column.Props, err = model.NewRawMessage(model.FloatRange{Min: 1, Max: 1000}); err != nil {
			return model.Column{}, fmt.Errorf("creating props for float range: %w", err)
		}

	case "timestamp", "timestamptz":
		column.Range = "timestamp"
		if column.Props, err = model.NewRawMessage(model.TimestampRange{
			Min:    time.Now().Add(-time.Hour * 87600).Truncate(time.Hour * 24), // 10 yeras
			Max:    time.Now().Truncate(time.Hour * 24),
			Format: "2006-01-02T15:04:05Z",
		}); err != nil {
			return model.Column{}, fmt.Errorf("creating props for timestamp range: %w", err)
		}

	case "geometry":
		column.Range = "point"
		if column.Props, err = model.NewRawMessage(model.PointRange{
			Lat:        51.542235,
			Lon:        -0.147515,
			DistanceKM: 20,
		}); err != nil {
			return model.Column{}, fmt.Errorf("creating props for point range: %w", err)
		}

	default:
		return model.Column{}, fmt.Errorf("invalid type %q", c.DataType)
	}

	return column, nil
}

func valueForTextColumn(c columnDefinition) string {
	name := strings.ToLower(c.ColumnName)

	switch {
	case strings.Contains(name, "email"):
		return "${email}"
	case strings.Contains(name, "name"):
		return "${name}"
	default:
		return "${COMPLETE}"
	}
}
