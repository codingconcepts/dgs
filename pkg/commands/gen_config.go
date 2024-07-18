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
	CharMaxLength    *int64
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
		if err = rows.Scan(&d.TableName, &d.ColumnName, &d.Default, &d.Nullable, &d.CharMaxLength, &d.DataType, &d.UserDefintedType, &d.ForeignKey); err != nil {
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

			column, ok, err := createRegularColumn(c)
			if err != nil {
				return nil, fmt.Errorf("creating column: %w", err)
			}

			if ok {
				table.Columns = append(table.Columns, column)
				continue
			}

			// Ignore unsupported columns.
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

func createRegularColumn(c columnDefinition) (model.Column, bool, error) {
	var err error
	var ok bool

	column := model.Column{
		Name: c.ColumnName,
	}

	// Handle array types.
	if strings.HasPrefix(c.DataType, "_") {
		column.Array, ok = valueForArrayColumn(c)
		if ok {
			if column.Props, err = model.NewRawMessage(model.IntRange{Min: 1, Max: 10}); err != nil {
				return model.Column{}, false, fmt.Errorf("creating props for array: %w", err)
			}

			return column, true, nil
		}

		return model.Column{}, false, nil
	}

	maxCharLen := int64(255)
	if c.CharMaxLength != nil {
		maxCharLen = *c.CharMaxLength
	}

	switch c.DataType {
	case "uuid":
		column.Value = "${uuid}"

	case "text", "varchar":
		column.Value, ok = valueForTextColumn(c)
		if !ok {
			column.Range = "string"
			if column.Props, err = model.NewRawMessage(model.IntRange{Min: 1, Max: maxCharLen}); err != nil {
				return model.Column{}, false, fmt.Errorf("creating props for string range: %w", err)
			}
		}

	case "inet":
		column.Value = "${ipv4_address}"

	case "bool":
		column.Value = "${bool}"

	case "int2":
		column.Value = "${int16}"

	case "int4":
		column.Value = "${int32}"

	case "oid":
		column.Value = "${uint32}"

	// case "bit":
	// 	column.Range = "bit"
	// 	if column.Props, err = model.NewRawMessage(model.IntRange{Min: maxCharLen, Max: maxCharLen}); err != nil {
	// 		return model.Column{}, false, fmt.Errorf("creating props for bit range: %w", err)
	// 	}

	// case "varbit":
	// 	column.Range = "bit"
	// 	if column.Props, err = model.NewRawMessage(model.IntRange{Min: 1, Max: maxCharLen}); err != nil {
	// 		return model.Column{}, false, fmt.Errorf("creating props for varbit range: %w", err)
	// 	}

	case "bytea":
		column.Range = "bytes"
		if column.Props, err = model.NewRawMessage(model.IntRange{Min: 1, Max: maxCharLen}); err != nil {
			return model.Column{}, false, fmt.Errorf("creating props for bytes range: %w", err)
		}

	case "int8":
		column.Range = "int"
		if column.Props, err = model.NewRawMessage(model.IntRange{Min: 1, Max: 1000000}); err != nil {
			return model.Column{}, false, fmt.Errorf("creating props for int range: %w", err)
		}

	case "numeric", "float4", "float8":
		column.Range = "float"
		if column.Props, err = model.NewRawMessage(model.FloatRange{Min: 1, Max: 1000}); err != nil {
			return model.Column{}, false, fmt.Errorf("creating props for float range: %w", err)
		}

	case "time", "timetz":
		column.Range = "timestamp"
		if column.Props, err = valueForTimeColumn("15:04:05"); err != nil {
			return model.Column{}, false, fmt.Errorf("creating props for timestamp range: %w", err)
		}

	case "timestamp", "timestamptz":
		column.Range = "timestamp"
		if column.Props, err = valueForTimeColumn("2006-01-02T15:04:05Z"); err != nil {
			return model.Column{}, false, fmt.Errorf("creating props for timestamp range: %w", err)
		}

	case "date":
		column.Range = "timestamp"
		if column.Props, err = valueForTimeColumn("2006-01-02"); err != nil {
			return model.Column{}, false, fmt.Errorf("creating props for timestamp range: %w", err)
		}

	case "interval":
		column.Range = "interval"
		if column.Props, err = model.NewRawMessage(model.IntervalRange{Min: time.Second, Max: time.Hour * 24}); err != nil {
			return model.Column{}, false, fmt.Errorf("creating props for interval range: %w", err)
		}

	case "geometry":
		column.Range = "point"
		if column.Props, err = model.NewRawMessage(model.PointRange{
			Lat:        51.542235,
			Lon:        -0.147515,
			DistanceKM: 20,
		}); err != nil {
			return model.Column{}, false, fmt.Errorf("creating props for point range: %w", err)
		}

	default:
		return model.Column{}, false, nil
	}

	return column, true, nil
}

func valueForTimeColumn(format string) (*model.RawMessage, error) {
	return model.NewRawMessage(model.TimestampRange{
		Min:    time.Now().Add(-time.Hour * 87600).Truncate(time.Hour * 24), // 10 yeras
		Max:    time.Now().Truncate(time.Hour * 24),
		Format: format,
	})
}

func valueForTextColumn(c columnDefinition) (string, bool) {
	name := strings.ToLower(c.ColumnName)

	switch {
	case strings.Contains(name, "email"):
		return "${email}", true
	case strings.Contains(name, "name"):
		return "${name}", true
	default:
		return "", false
	}
}

func valueForArrayColumn(c columnDefinition) (string, bool) {
	itemType := strings.TrimPrefix(c.DataType, "_")

	switch strings.ToLower(itemType) {
	case "text":
		return "${fruit}", true
	case "int8":
		return "${http_status_code_simple}", true
	}

	return "", false
}
