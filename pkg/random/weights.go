package random

import (
	"fmt"

	"github.com/samber/lo"
)

func DefaultWeights(items int) []int64 {
	result := make([]int64, items)
	baseValue := int64(100 / items)
	remainder := 100 % items

	for i := range result {
		result[i] = baseValue
	}

	for i := range remainder {
		result[i]++
	}

	return result
}

func BuildWeightedItems[T any](values []T, weights []int64) (weightedItems[T], error) {
	if len(values) != len(weights) {
		return weightedItems[T]{}, fmt.Errorf("set values and weights need to be the same")
	}

	items := make([]weightedItem[T], len(values))
	for i, v := range values {
		items[i] = weightedItem[T]{
			Value:  v,
			Weight: weights[i],
		}
	}

	return makeWeightedItems(items), nil
}

type weightedItem[T any] struct {
	Value  T
	Weight int64
}

type weightedItems[T any] struct {
	items       []weightedItem[T]
	totalWeight int64
}

func makeWeightedItems[T any](items []weightedItem[T]) weightedItems[T] {
	wi := weightedItems[T]{
		items: items,
	}

	wi.totalWeight = lo.SumBy(items, func(wi weightedItem[T]) int64 {
		return wi.Weight
	})

	return wi
}

func (wi weightedItems[T]) Choose() T {
	randomWeight := Int(1, wi.totalWeight)
	for _, i := range wi.items {
		randomWeight -= i.Weight
		if randomWeight <= 0 {
			return i.Value
		}
	}

	// Return zero value of T if no item is selected.
	var zero T
	return zero
}
