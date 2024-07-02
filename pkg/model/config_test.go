package model

import (
	"errors"
	"testing"

	"github.com/codingconcepts/dgs/pkg/test"
	"github.com/stretchr/testify/assert"
)

func TestParseConfig_ValidYAML(t *testing.T) {
	yaml := `tables:
  - name: ta
    rows: 10
    columns:
      - name: ca
        value: ${uuid}
      - name: cb
        value: ${email}

  - name: tb
    rows: 20
    columns:
      - name: cc
        value: ${name}
      - name: cd
        range: timestamp
        format: 2006-01-02T15:04:05Z
        min: 2020-01-01T00:00:00Z
        max: 2024-01-01T00:00:00Z`

	config, err := ParseConfig(yaml, test.NewNilLogger())
	assert.NoError(t, err)

	exp := Config{
		Tables: []Table{
			{
				Name: "ta",
				Rows: 10,
				Columns: []Column{
					{
						Name:  "ca",
						Mode:  "value",
						Value: "${uuid}",
					},
					{
						Name:  "cb",
						Mode:  "value",
						Value: "${email}",
					},
				},
			},
			{
				Name: "tb",
				Rows: 20,
				Columns: []Column{
					{
						Name:  "cc",
						Mode:  "value",
						Value: "${name}",
					},
					{
						Name:   "cd",
						Mode:   "range",
						Range:  "timestamp",
						Format: "2006-01-02T15:04:05Z",
						Min:    "2020-01-01T00:00:00Z",
						Max:    "2024-01-01T00:00:00Z",
					},
				},
			},
		},
	}

	assert.Equal(t, exp, config)
}

func TestParseColumn(t *testing.T) {
	tests := []struct {
		name         string
		column       Column
		expectedMode ColumnType
		expError     error
	}{
		{
			name: "value type",
			column: Column{
				Name:  "col",
				Value: "${uuid}",
			},
			expectedMode: ColumnTypeValue,
		},
		{
			name: "range type",
			column: Column{
				Name:  "col",
				Range: "timestamp",
			},
			expectedMode: ColumnTypeRange,
		},
		{
			name: "ref type",
			column: Column{
				Name: "col",
				Ref:  "table.column",
			},
			expectedMode: ColumnTypeRef,
		},
		{
			name: "set type",
			column: Column{
				Name: "col",
				Set:  []string{"a", "b", "c"},
			},
			expectedMode: ColumnTypeSet,
		},
		{
			name: "missing mode",
			column: Column{
				Name: "col",
			},
			expError: errors.New("missing value, range, ref, or set for column"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			table := &Table{
				Columns: []Column{tt.column},
			}

			err := parseColumn(table, 0)
			if tt.expError != nil {
				assert.Equal(t, tt.expError, err)
				return
			}

			assert.Equal(t, tt.expectedMode, table.Columns[0].Mode)
		})
	}
}
