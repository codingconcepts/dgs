package test

import (
	"testing"

	"gopkg.in/yaml.v3"
)

func CreateNode(t *testing.T, val any) yaml.Node {
	b, err := yaml.Marshal(val)
	if err != nil {
		t.Fatalf("error marshalling node: %v", err)
	}

	var node yaml.Node
	if err = yaml.Unmarshal(b, &node); err != nil {
		t.Fatalf("error unmarshalling node: %v", err)
	}

	return node
}
