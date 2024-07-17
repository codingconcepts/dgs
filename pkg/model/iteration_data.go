package model

import (
	"fmt"

	"github.com/samber/lo"
)

type IterationData struct {
	data map[string][]any
}

func NewIterationData() *IterationData {
	return &IterationData{
		data: map[string][]any{},
	}
}

func (d *IterationData) AddData(rows [][]any, table, column string, index int) {
	var values []any
	for _, row := range rows {
		values = append(values, row[index])
	}

	refKey := fmt.Sprintf("%s.%s", table, column)
	d.data[refKey] = values
}

func (d *IterationData) GetValue(ref string) any {
	return lo.Sample(d.data[ref])
}

func (d *IterationData) GetValues(ref string) []any {
	return d.data[ref]
}
