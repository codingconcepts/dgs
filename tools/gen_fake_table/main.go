package main

import (
	"fmt"

	"github.com/codingconcepts/dgs/pkg/random"
)

func main() {
	fmt.Println("| Fake function | Example |")
	fmt.Println("| ------------- | ------- |")
	for k, v := range random.Replacements {
		fmt.Printf("| %s | %v |\n", k, v())
	}
}
