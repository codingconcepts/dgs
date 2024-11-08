package query

import (
	"fmt"
	"strings"

	"github.com/codingconcepts/dgs/pkg/model"

	"github.com/samber/lo"
)

func BuildInsert(table model.Table, rows [][]any, insertMode model.InsertMode) (string, error) {
	var b model.ErrBuilder

	columnNames := lo.Map(table.Columns, func(c model.Column, i int) string {
		return c.Name
	})

	b.WriteString(
		"%s INTO %s (%s) VALUES ",
		lo.Ternary(insertMode == model.InsertModeInsert || insertMode == model.InsertModeConflict, "INSERT", "UPSERT"),
		table.Name,
		strings.Join(columnNames, ","),
	)

	argIndex := 1
	for i, row := range rows {
		columnValues, err := valuePlaceholders(len(row), argIndex)
		if err != nil {
			return "", fmt.Errorf("generating value placeholders: %w", err)
		}

		b.WriteString("(%s)", columnValues)

		if i < len(rows)-1 {
			b.WriteString(",")
		}

		argIndex += len(table.Columns)
	}

	if insertMode == model.InsertModeConflict {
		b.WriteString(" ON CONFLICT DO NOTHING")
	}

	if err := b.Error(); err != nil {
		return "", err
	}

	return b.String(), nil
}

func valuePlaceholders(count, start int) (string, error) {
	var b model.ErrBuilder

	total := start + count - 1
	for i := start; i <= total; i++ {
		b.WriteString(fmt.Sprintf("$%d", i))
		if i < total {
			b.WriteString(",")
		}
	}

	if err := b.Error(); err != nil {
		return "", err
	}

	return b.String(), nil
}
