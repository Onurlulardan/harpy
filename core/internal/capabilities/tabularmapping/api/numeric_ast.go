package tabularmapping

import (
	"encoding/json"
	"regexp"
)

var decimalPattern = regexp.MustCompile(`^-?(0|[1-9][0-9]*)(\.[0-9]+)?$`)

type numericNode struct {
	Kind      string          `json:"kind"`
	Value     json.RawMessage `json:"value,omitempty"`
	Source    string          `json:"source,omitempty"`
	Arguments []numericNode   `json:"arguments,omitempty"`
}

func validateNumericNode(node numericNode, depth int) (int, error) {
	if depth > MaximumNumericDepth {
		return 0, ErrInvalid
	}
	switch node.Kind {
	case "literal":
		if node.Source != "" || node.Arguments != nil || !validNumericLiteral(node.Value) {
			return 0, ErrInvalid
		}
	case "source":
		if !validSource(node.Source) || node.Value != nil || node.Arguments != nil {
			return 0, ErrInvalid
		}
	case "negate":
		if node.Source != "" || node.Value != nil || len(node.Arguments) != 1 {
			return 0, ErrInvalid
		}
	case "add", "subtract", "multiply", "divide":
		if node.Source != "" || node.Value != nil || len(node.Arguments) != 2 {
			return 0, ErrInvalid
		}
	default:
		return 0, ErrInvalid
	}
	nodes := 1
	for _, child := range node.Arguments {
		count, err := validateNumericNode(child, depth+1)
		if err != nil || nodes+count > MaximumNumericNodesPerField {
			return 0, ErrInvalid
		}
		nodes += count
	}
	return nodes, nil
}

func validNumericLiteral(raw json.RawMessage) bool {
	text, ok := numericText(raw)
	return ok && len(text) <= MaximumSourceBytes && decimalPattern.MatchString(text)
}
