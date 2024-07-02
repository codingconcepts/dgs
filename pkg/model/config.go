package model

import (
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
	Workers int     `yaml:"workers"`
	Batch   int     `yaml:"batch"`
	Tables  []Table `yaml:"tables"`
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
			if table.Columns[i].Value != "" {
				table.Columns[i].Mode = ColumnTypeValue
				logger.Info().Msgf("column %s is a %q type", table.Columns[i].Name, table.Columns[i].Mode)
				continue
			}

			if table.Columns[i].Range != "" {
				table.Columns[i].Mode = ColumnTypeRange
				logger.Info().Msgf("column %s is a %q type", table.Columns[i].Name, table.Columns[i].Mode)
				continue
			}

			if table.Columns[i].Ref != "" {
				table.Columns[i].Mode = ColumnTypeRef
				logger.Info().Msgf("column %s is a %q type", table.Columns[i].Name, table.Columns[i].Mode)
				continue
			}

			if table.Columns[i].Set != nil {
				table.Columns[i].Mode = ColumnTypeSet
				logger.Info().Msgf("column %s is a %q type", table.Columns[i].Name, table.Columns[i].Mode)
				continue
			}
		}
	}

	return config, nil
}
