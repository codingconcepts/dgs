package commands

import (
	"errors"
	"testing"

	"github.com/codingconcepts/dgs/pkg/model"
	"github.com/stretchr/testify/assert"
)

func TestCalculateIterations(t *testing.T) {
	cases := []struct {
		name   string
		tables []model.Table
		batch  int
		exp    map[string]int
		expErr error
	}{
		{
			name: "incrementing",
			tables: []model.Table{
				{Name: "a", Rows: 1000},
				{Name: "b", Rows: 2000},
				{Name: "c", Rows: 4000},
			},
			batch: 100,
			exp: map[string]int{
				"a": 1,
				"b": 2,
				"c": 4,
			},
		},
		{
			name: "decrementing",
			tables: []model.Table{
				{Name: "a", Rows: 4000},
				{Name: "b", Rows: 2000},
				{Name: "c", Rows: 1000},
			},
			batch: 100,
			exp: map[string]int{
				"a": 4,
				"b": 2,
				"c": 1,
			},
		},
		{
			name: "batch equal to smallest table",
			tables: []model.Table{
				{Name: "a", Rows: 4000},
				{Name: "b", Rows: 2000},
				{Name: "c", Rows: 1000},
			},
			batch: 1000,
			exp: map[string]int{
				"a": 4,
				"b": 2,
				"c": 1,
			},
		},
		{
			name: "larger batch than smallest table",
			tables: []model.Table{
				{Name: "a", Rows: 4000},
				{Name: "b", Rows: 2000},
				{Name: "c", Rows: 100},
			},
			batch:  1000,
			expErr: errors.New("smallest table must have at least 1000 rows"),
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			act, err := calculateIterations(c.tables, c.batch)
			assert.Equal(t, c.expErr, err)
			if err != nil {
				return
			}

			assert.Equal(t, c.exp, act)
		})
	}
}
