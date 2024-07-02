package model

import (
	"fmt"

	"github.com/rs/zerolog"
	"gopkg.in/yaml.v2"
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
	Name    string   `yaml:"name"`
	Rows    int      `yaml:"rows"`
	Columns []Column `yaml:"columns"`
}

type Column struct {
	Name   string `yaml:"name"`
	Mode   ColumnType
	Value  string   `yaml:"value"`
	Range  string   `yaml:"range"`
	Min    string   `yaml:"min,omitempty"`
	Max    string   `yaml:"max,omitempty"`
	Length int      `yaml:"length"`
	Format string   `yaml:"format"`
	Ref    string   `yaml:"ref,omitempty"`
	Set    []string `yaml:"set,omitempty"`
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
