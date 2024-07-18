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
	ColumnTypeInc   ColumnType = "inc"
	ColumnTypeArray ColumnType = "array"
)

type Config struct {
	Tables []Table `yaml:"tables"`
}

type Table struct {
	Name    string   `yaml:"name"`
	Rows    int      `yaml:"rows"`
	Columns []Column `yaml:"columns"`

	RefColumns []string `yaml:"-"`
}

type Column struct {
	Name  string      `yaml:"name"`
	Mode  ColumnType  `yaml:"-"`
	Value string      `yaml:"value,omitempty"`
	Array string      `yaml:"array,omitempty"`
	Range string      `yaml:"range,omitempty"`
	Props *RawMessage `yaml:"props,omitempty"`
	Ref   string      `yaml:"ref,omitempty"`
	Set   []string    `yaml:"set,omitempty"`
	Inc   int64       `yaml:"inc,omitempty"`

	NextID Sequence `yaml:"-"`
}

type IntRange struct {
	Min int64 `yaml:"min"`
	Max int64 `yaml:"max"`
}

type FloatRange struct {
	Min float64 `yaml:"min"`
	Max float64 `yaml:"max"`
}

type TimestampRange struct {
	Min    time.Time `yaml:"min"`
	Max    time.Time `yaml:"max"`
	Format string    `yaml:"format"`
}

type IntervalRange struct {
	Min time.Duration `yaml:"min"`
	Max time.Duration `yaml:"max"`
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
			logger.Info().
				Str("name", table.Columns[i].Name).
				Str("mode", string(table.Columns[i].Mode)).
				Msgf("parsed column")
		}
	}

	// Mark tables that are dependencies on others.
	markDependencies(&config)

	return config, nil
}

func markDependencies(c *Config) {
	tableMap := make(map[string]*Table)
	for i := range c.Tables {
		tableMap[c.Tables[i].Name] = &c.Tables[i]
	}

	for i := range c.Tables {
		for _, column := range c.Tables[i].Columns {
			if column.Ref == "" {
				continue
			}

			refParts := strings.Split(column.Ref, ".")
			if len(refParts) != 2 {
				continue
			}

			if table, exists := tableMap[refParts[0]]; exists {
				table.RefColumns = append(table.RefColumns, refParts[1])
			}
		}
	}
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
	case table.Columns[i].Inc != 0:
		table.Columns[i].Mode = ColumnTypeInc
		table.Columns[i].NextID = Inc(table.Columns[i].Inc)
	case table.Columns[i].Array != "":
		table.Columns[i].Mode = ColumnTypeArray
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
