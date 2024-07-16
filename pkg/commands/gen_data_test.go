package commands

import (
	"errors"
	"testing"

	"github.com/codingconcepts/dgs/pkg/model"
	"github.com/stretchr/testify/assert"
)

func TestCalculateIterations(t *testing.T) {
	cases := []struct {
		name          string
		tables        []model.Table
		batch         int
		expIterations map[string]int
		expLoop       int
		expError      error
	}{
		{
			name: "batch lt smallest table and divisible",
			tables: []model.Table{
				{Name: "a", Rows: 1000},
				{Name: "b", Rows: 2000},
				{Name: "c", Rows: 4000},
			},
			batch: 100,
			expIterations: map[string]int{
				"a": 1,
				"b": 2,
				"c": 4,
			},
			expLoop: 10,
		},
		{
			name: "batch lt smallest table and indivisible",
			tables: []model.Table{
				{Name: "a", Rows: 1000},
				{Name: "b", Rows: 2000},
				{Name: "c", Rows: 4000},
			},
			batch: 67,
			expIterations: map[string]int{
				"a": 1,
				"b": 2,
				"c": 4,
			},
			expLoop: 14,
		},
		{
			name: "batch eq smallest table",
			tables: []model.Table{
				{Name: "a", Rows: 1000},
				{Name: "b", Rows: 2000},
				{Name: "c", Rows: 4000},
			},
			batch: 1000,
			expIterations: map[string]int{
				"a": 1,
				"b": 2,
				"c": 4,
			},
			expLoop: 1,
		},
		{
			name: "batch gt smallest table",
			tables: []model.Table{
				{Name: "a", Rows: 1000},
				{Name: "b", Rows: 2000},
				{Name: "c", Rows: 4000},
			},
			batch:    10000,
			expError: errors.New("batch size should be <= 1000"),
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			sut := &DataGenerator{
				batch: c.batch,
				config: model.Config{
					Tables: c.tables,
				},
			}

			actLoop, actIterations, actErr := sut.calculateIterations()
			assert.Equal(t, c.expError, actErr)
			if actErr != nil {
				return
			}

			assert.Equal(t, c.expIterations, actIterations)
			assert.Equal(t, c.expLoop, actLoop)
		})
	}
}
