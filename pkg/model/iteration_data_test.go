package model

import (
	"testing"

	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
)

func TestIterationData_Add(t *testing.T) {
	rows := [][]any{
		{1, "a"},
		{2, "b"},
		{3, "c"},
	}

	exp := map[string][]any{
		"table.id": {1, 2, 3},
	}

	sut := NewIterationData()
	sut.AddData(rows, "table", "id", 0)

	assert.Equal(t, exp, sut.data)
}

func TestIterationData_Sample(t *testing.T) {
	sut := &IterationData{
		data: map[string][]any{
			"table.id": {1, 2, 3},
		},
	}

	sample := sut.GetValue("table.id")

	assert.True(t, lo.Contains(sut.data["table.id"], sample))
}
