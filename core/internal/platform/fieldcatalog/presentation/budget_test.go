package presentation

import (
	"errors"
	"fmt"
	"testing"
)

func TestNodeCountBudgetBoundary(t *testing.T) {
	for _, test := range []struct {
		children int
		valid    bool
	}{{maximumNodes - 1, true}, {maximumNodes, false}} {
		document := testDocument("text", "auto")
		nodes := make([]Node, test.children)
		for index := range nodes {
			nodes[index] = Node{
				ID:   fmt.Sprintf("018f47f4-7b1b-7cc0-8000-%012x", index+0x100),
				Kind: "divider", Span: 1, Translations: []Translation{},
			}
		}
		document.Surfaces[0].Root.Children = nodes
		err := Validate(document, testCatalog(t, FieldEditable))
		if test.valid && err != nil || !test.valid && !errors.Is(err, ErrInvalid) {
			t.Fatalf("children=%d valid=%v error=%v", test.children, test.valid, err)
		}
	}
}

func TestNodeDepthBudgetBoundary(t *testing.T) {
	for _, test := range []struct {
		depth int
		valid bool
	}{{maximumDepth, true}, {maximumDepth + 1, false}} {
		document := testDocument("text", "auto")
		node := document.Surfaces[0].Root.Children[0]
		for level := 2; level <= test.depth; level++ {
			node = Node{
				ID:   fmt.Sprintf("018f47f4-7b1b-7cc0-8000-%012x", level+0x100),
				Kind: "container", Variant: "stack", Span: 1,
				Translations: []Translation{}, Children: []Node{node},
			}
		}
		document.Surfaces[0].Root = node
		err := Validate(document, testCatalog(t, FieldEditable))
		if test.valid && err != nil || !test.valid && !errors.Is(err, ErrInvalid) {
			t.Fatalf("depth=%d valid=%v error=%v", test.depth, test.valid, err)
		}
	}
}
