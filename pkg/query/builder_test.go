package query

import (
	"testing"

	"github.com/codingconcepts/dgs/pkg/model"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
)

func TestBuildInsert(t *testing.T) {
	cases := []struct {
		name         string
		tableName    string
		columnNames  []string
		rows         [][]any
		expStatement string
	}{
		{
			name:         "1x3",
			tableName:    "t",
			columnNames:  []string{"a"},
			rows:         [][]any{{1}, {2}, {3}},
			expStatement: `INSERT INTO t (a) VALUES ($1),($2),($3)`,
		},
		{
			name:         "2x3",
			tableName:    "t",
			columnNames:  []string{"a", "b", "c"},
			rows:         [][]any{{1, 2, 3}, {4, 5, 6}},
			expStatement: `INSERT INTO t (a,b,c) VALUES ($1,$2,$3),($4,$5,$6)`,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			table := model.Table{
				Name: c.tableName,
				Columns: lo.Map(c.columnNames, func(n string, i int) model.Column {
					return model.Column{
						Name: n,
					}
				}),
			}

			actStatement, err := BuildInsert(table, c.rows)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			assert.Equal(t, c.expStatement, actStatement)
		})
	}
}
