package main

import (
	"fmt"
	"sort"

	"github.com/codingconcepts/dgs/pkg/random"
)

func main() {
	var entries []entry
	for k, v := range random.Replacements {
		entries = append(entries, entry{
			name:  k,
			value: v(),
		})
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].name < entries[j].name
	})

	fmt.Println("| Fake function | Example |")
	fmt.Println("| ------------- | ------- |")
	for _, entry := range entries {
		fmt.Printf("| %s | %v |\n", entry.name, entry.value)
	}
}

type entry struct {
	name  string
	value any
}
