package commands

import (
	"context"
	"fmt"
	"time"

	"github.com/codingconcepts/dgs/pkg/model"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/samber/lo"
)

// GenerateConfig creates a config object
func GenerateConfig(db *pgxpool.Pool, schema string) (model.Config, error) {
	columns, err := fetchColumnDefinitions(db, schema)
	if err != nil {
		return model.Config{}, fmt.Errorf("fetching column definitions: %w", err)
	}

	tables, err := toConfigs(columns)
	if err != nil {
		return model.Config{}, fmt.Errorf("converting column defintions to config: %w", err)
	}

	return model.Config{
		Tables: tables,
	}, nil
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
	const stmt = `WITH
									columns_info AS (
										SELECT 
											table_name,
											column_name,
											ordinal_position,
											column_default,
											is_nullable,
											udt_name AS data_type
										FROM information_schema.columns
										WHERE table_schema = $1
									),
									foreign_keys_info AS (
										SELECT
											tc.constraint_name,
											tc.table_name AS fk_table,
											kcu.column_name AS fk_column,
											ccu.table_name AS pk_table,
											ccu.column_name AS pk_column
										FROM information_schema.table_constraints AS tc
										JOIN information_schema.key_column_usage AS kcu ON tc.constraint_name = kcu.constraint_name
										JOIN information_schema.constraint_column_usage AS ccu ON ccu.constraint_name = tc.constraint_name
										WHERE tc.constraint_type = 'FOREIGN KEY'
										AND tc.table_schema = $1
									),
									user_defined_types AS (
										SELECT 
											t.typname AS type_name,
											t.typtype AS type_type,
											array_agg(e.enumlabel ORDER BY e.enumsortorder) AS enum_labels
										FROM  pg_type t
										JOIN pg_namespace n ON t.typnamespace = n.oid
										LEFT JOIN pg_enum e ON t.oid = e.enumtypid
										WHERE n.nspname = $1
										GROUP BY t.typname, t.typtype
									)
								SELECT 
									c.table_name,
									c.column_name,
									c.column_default,
									c.is_nullable,
									CASE
										WHEN udt.type_type = 'e' THEN 'enum'
										ELSE c.data_type
									END AS "data_type",
									CASE 
										WHEN udt.type_type = 'e' THEN udt.enum_labels
										ELSE NULL
									END AS user_defined_type,
									CASE
										WHEN fk.pk_table IS NOT NULL THEN fk.pk_table || '.' || fk.pk_column
									END AS "fk"
								FROM columns_info AS c
								LEFT JOIN foreign_keys_info AS fk
								ON c.table_name = fk.fk_table AND c.column_name = fk.fk_column
								LEFT JOIN user_defined_types AS udt
								ON c.data_type = udt.type_name
								ORDER BY c.table_name, c.ordinal_position`

	rows, err := db.Query(context.Background(), stmt, schema)
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

func toConfigs(definitions []columnDefinition) ([]model.Table, error) {
	groups := lo.GroupBy(definitions, func(d columnDefinition) string {
		return d.TableName
	})

	var tables []model.Table
	for _, t := range groups {
		table := model.Table{
			Name: t[0].TableName,
			Rows: 100000,
		}

		for _, c := range t {
			column := model.Column{
				Name: c.ColumnName,
			}

			if c.ForeignKey != nil {
				column.Mode = model.ColumnTypeRef
				column.Ref = *c.ForeignKey
				table.Columns = append(table.Columns, column)
				continue
			}

			if c.DataType == "enum" {
				if c.UserDefintedType == nil {
					return nil, fmt.Errorf("missing values for enum column %q", c.ColumnName)
				}

				column.Mode = model.ColumnTypeSet
				column.Set = *c.UserDefintedType
				table.Columns = append(table.Columns, column)
				continue
			}

			var err error
			switch c.DataType {
			case "uuid":
				column.Mode = model.ColumnTypeValue
				column.Value = "${uuid}"
			case "text":
				column.Mode = model.ColumnTypeValue
				column.Value = "${COMPLETE}"
			case "int8":
				column.Mode = model.ColumnTypeRange
				column.Range = "int"
				if column.Props, err = model.NewRawMessage(model.IntRange{Min: 1, Max: 1000000}); err != nil {
					return nil, fmt.Errorf("creating props for int range: %w", err)
				}
			case "numeric":
				column.Mode = model.ColumnTypeRange
				column.Range = "float"
				if column.Props, err = model.NewRawMessage(model.FloatRange{Min: 1, Max: 1000}); err != nil {
					return nil, fmt.Errorf("creating props for float range: %w", err)
				}
			case "timestamp", "timestamptz":
				column.Mode = model.ColumnTypeRange
				column.Range = "timestamp"
				if column.Props, err = model.NewRawMessage(model.TimestampRange{
					Min:    time.Now().Add(-time.Hour * 87600).Truncate(time.Hour * 24), // 10 yeras
					Max:    time.Now().Truncate(time.Hour * 24),
					Format: "2006-01-02T15:04:05Z",
				}); err != nil {
					return nil, fmt.Errorf("creating props for timestamp range: %w", err)
				}
			default:
				return nil, fmt.Errorf("invalid type %q", c.DataType)
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
