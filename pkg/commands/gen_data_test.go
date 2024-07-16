package commands

import (
	"testing"

	"github.com/codingconcepts/dgs/pkg/model"
	"github.com/stretchr/testify/assert"
)

func TestCalculateIterations(t *testing.T) {
	cases := []struct {
		name    string
		tables  []model.Table
		batch   int
		workers int
		exp     map[string]iteration
	}{
		{
			name: "small batch 1 worker",
			tables: []model.Table{
				{Name: "a", Rows: 1000},
				{Name: "b", Rows: 2000},
				{Name: "c", Rows: 4000},
			},
			batch:   100,
			workers: 1,
			exp: map[string]iteration{
				"a": {batch: 100, times: 10},
				"b": {batch: 100, times: 20},
				"c": {batch: 100, times: 40},
			},
		},
		{
			name: "small batch 4 workers",
			tables: []model.Table{
				{Name: "a", Rows: 1000},
				{Name: "b", Rows: 2000},
				{Name: "c", Rows: 4000},
			},
			batch:   100,
			workers: 1,
			exp: map[string]iteration{
				"a": {batch: 100, times: 10},
				"b": {batch: 100, times: 20},
				"c": {batch: 100, times: 40},
			},
		},
		{
			name: "equal batch 1 worker",
			tables: []model.Table{
				{Name: "a", Rows: 1000},
				{Name: "b", Rows: 2000},
				{Name: "c", Rows: 4000},
			},
			batch:   1000,
			workers: 1,
			exp: map[string]iteration{
				"a": {batch: 1000, times: 1},
				"b": {batch: 1000, times: 2},
				"c": {batch: 1000, times: 4},
			},
		},
		{
			name: "equal batch 4 workers",
			tables: []model.Table{
				{Name: "a", Rows: 1000},
				{Name: "b", Rows: 2000},
				{Name: "c", Rows: 4000},
			},
			batch:   1000,
			workers: 4,
			exp: map[string]iteration{
				"a": {batch: 250, times: 1},
				"b": {batch: 1000, times: 1},
				"c": {batch: 1000, times: 1},
			},
		},
		{
			name: "large batch 1 worker",
			tables: []model.Table{
				{Name: "a", Rows: 1000},
				{Name: "b", Rows: 2000},
				{Name: "c", Rows: 4000},
			},
			batch:   10000,
			workers: 1,
			exp: map[string]iteration{
				"a": {batch: 1000, times: 1},
				"b": {batch: 2000, times: 1},
				"c": {batch: 4000, times: 1},
			},
		},
		{
			name: "large batch 4 workers",
			tables: []model.Table{
				{Name: "a", Rows: 1000},
				{Name: "b", Rows: 2000},
				{Name: "c", Rows: 4000},
			},
			batch:   10000,
			workers: 4,
			exp: map[string]iteration{
				"a": {batch: 250, times: 1},
				"b": {batch: 500, times: 1},
				"c": {batch: 1000, times: 1},
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			sut := &DataGenerator{
				batch:   c.batch,
				workers: c.workers,
				config: model.Config{
					Tables: c.tables,
				},
			}

			actIterations := sut.calculateIterations()
			assert.Equal(t, c.exp, actIterations)
		})
	}
}
