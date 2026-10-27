package executor

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/planner"
)

func TestAggregateResultIsImmutableRedactedAndSealed(t *testing.T) {
	fixture := newAggregateFixture(t, false)
	batch := consumedAggregateFixture(t, fixture)
	driver, _ := readyAggregateDriver(t, batch, [][]any{{int64(0), nil}}, "1")
	result, err := runAggregate(context.Background(), driver, batch, planner.HardPolicy())
	if err != nil {
		t.Fatal(err)
	}
	groups, _ := result.Groups()
	groups[0][2] ^= 1
	again, err := result.Groups()
	if err != nil || string(again[0]) != `{"key":{},"measures":{"row_count":"0","revenue":null}}` {
		t.Fatal("aggregate result accessor exposed mutable state")
	}
	if fmt.Sprint(result) != redacted || result.LogValue().String() != redacted {
		t.Fatal("aggregate result formatting leaked")
	}
	if _, err := json.Marshal(result); !errors.Is(err, ErrDirectMarshal) {
		t.Fatalf("direct marshal error=%v", err)
	}
	tampered := result
	tampered.stateSeal[0] ^= 1
	if !errors.Is(tampered.Validate(), ErrContract) {
		t.Fatal("tampered aggregate result validated")
	}
}
