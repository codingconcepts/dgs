package model

import (
	"fmt"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"gopkg.in/yaml.v3"
)

type ColumnType string

const (
	ColumnTypeValue ColumnType = "value"
	ColumnTypeRange ColumnType = "range"
	ColumnTypeRef   ColumnType = "ref"
	ColumnTypeSet   ColumnType = "set"
)

type Config struct {
	Tables []Table `yaml:"tables"`
}

type Table struct {
	Name        string   `yaml:"name"`
	Rows        int      `yaml:"rows"`
	Columns     []Column `yaml:"columns"`
	Identifiers []string `yaml:"identifiers"`
}

type Column struct {
	Name  string `yaml:"name"`
	Mode  ColumnType
	Value string      `yaml:"value"`
	Range string      `yaml:"range"`
	Props *RawMessage `yaml:"props"`
	Ref   string      `yaml:"ref,omitempty"`
	Set   []string    `yaml:"set,omitempty"`
}

type IntRange struct {
	Min int64 `yaml:"min"`
	Max int64 `yaml:"max"`
}

type FloatRange struct {
	Min float64 `yaml:"min"`
	Max float64 `yaml:"max"`
}

type ByteRange struct {
	Min int64 `yaml:"min"`
	Max int64 `yaml:"max"`
}

type TimestampRange struct {
	Min    time.Time `yaml:"min"`
	Max    time.Time `yaml:"max"`
	Format string    `yaml:"format"`
}

type PointRange struct {
	Lat        float64 `yaml:"lat"`
	Lon        float64 `yaml:"lon"`
	DistanceKM float64 `yaml:"distance_km"`
}

func ParseConfig(yamlData string, logger zerolog.Logger) (Config, error) {
	var config Config
	err := yaml.Unmarshal([]byte(yamlData), &config)
	if err != nil {
		return Config{}, err
	}

	// Set mode and initialize any mode dependencies.
	for _, table := range config.Tables {
		logger.Info().Str("table", table.Name).Msg("parsing column types")

		for i := range table.Columns {
			if err = parseColumn(&table, i); err != nil {
				return Config{}, fmt.Errorf("parsing column %q: %w", table.Columns[i].Name, err)
			}
			logger.Info().Msgf("column %s is a %q type", table.Columns[i].Name, table.Columns[i].Mode)
		}
	}

	return config, nil
}

func parseColumn(table *Table, i int) error {
	switch {
	case table.Columns[i].Value != "":
		table.Columns[i].Mode = ColumnTypeValue
	case table.Columns[i].Range != "":
		table.Columns[i].Mode = ColumnTypeRange
	case table.Columns[i].Ref != "":
		table.Columns[i].Mode = ColumnTypeRef
	case table.Columns[i].Set != nil:
		table.Columns[i].Mode = ColumnTypeSet
	default:
		return fmt.Errorf("missing value, range, ref, or set for column")
	}
	return nil
}

// Sort tables by their inter-dependence.
func SortTables(tables []Table) ([]Table, error) {
	dependencies := make(map[string][]string)
	inDegree := make(map[string]int)

	// Initialize inDegree for all tables.
	for _, table := range tables {
		inDegree[table.Name] = 0
	}

	// Populate dependencies and in-degree map.
	for _, table := range tables {
		for _, col := range table.Columns {
			if col.Ref == "" {
				continue
			}

			refTable := strings.Split(col.Ref, ".")[0]
			dependencies[refTable] = append(dependencies[refTable], table.Name)
			inDegree[table.Name]++
		}
	}

	// Topological Sort using Kahn's Algorithm.
	var sortedTables []Table
	queue := []string{}

	// Start with tables that have no incoming edges (in-degree 0).
	for table, degree := range inDegree {
		if degree == 0 {
			queue = append(queue, table)
		}
	}

	for len(queue) > 0 {
		// Dequeue a table.
		currentTable := queue[0]
		queue = queue[1:]

		// Add to the sorted list.
		for _, table := range tables {
			if table.Name == currentTable {
				sortedTables = append(sortedTables, table)
				break
			}
		}

		// Decrease in-degree of dependent tables.
		for _, dependentTable := range dependencies[currentTable] {
			inDegree[dependentTable]--
			if inDegree[dependentTable] == 0 {
				queue = append(queue, dependentTable)
			}
		}
	}

	// Check for cyclic dependencies and fail if found.
	if len(sortedTables) != len(tables) {
		return nil, fmt.Errorf("cyclic dependency detected")
	}

	return sortedTables, nil
}
