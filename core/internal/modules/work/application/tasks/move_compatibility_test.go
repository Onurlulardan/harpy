package tasks

import (
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/modules/work/domain"
	"github.com/google/uuid"
)

func TestLegacyOrderNeighborsPreservesVisibleIndex(t *testing.T) {
	first, moving, last := domain.ID(uuid.New()), domain.ID(uuid.New()), domain.ID(uuid.New())
	ranked := []domain.RankedTask{{ID: first, Position: 1024},
		{ID: moving, Position: 2048}, {ID: last, Position: 3072}}
	tests := []struct {
		name   string
		order  int
		before *domain.ID
		after  *domain.ID
	}{
		{"negative becomes first", -1, nil, &first},
		{"zero becomes first", 0, nil, &first},
		{"middle uses adjacent neighbors", 1, &first, &last},
		{"large becomes last", 99, &last, nil},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := legacyOrderNeighbors(ranked, moving, test.order)
			if !sameOptionalID(got.Before, test.before) || !sameOptionalID(got.After, test.after) {
				t.Fatalf("neighbors = %#v, want before=%v after=%v", got, test.before, test.after)
			}
		})
	}
}

func sameOptionalID(left, right *domain.ID) bool {
	return left == nil && right == nil || left != nil && right != nil && *left == *right
}
