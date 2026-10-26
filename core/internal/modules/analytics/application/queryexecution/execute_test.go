package queryexecution

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestExecuteBoundsConcurrencyAndIsolatesFailure(t *testing.T) {
	var active, maximum atomic.Int32
	runner := runnerFunc(func(ctx context.Context, item Item) ([]byte, error) {
		current := active.Add(1)
		defer active.Add(-1)
		for current > maximum.Load() && !maximum.CompareAndSwap(maximum.Load(), current) {
		}
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(5 * time.Millisecond):
		}
		if item.QueryID == uuid.MustParse("019fa000-0000-7000-8000-000000000009") {
			return nil, errors.New("hidden")
		}
		return []byte(`{"items":[]}`), nil
	})
	request := Request{Items: make([]Item, 12)}
	for index := range request.Items {
		request.Items[index] = fixtureItem(index + 1)
	}

	result, err := Execute(context.Background(), request, runner)
	if err != nil || maximum.Load() > MaximumConcurrency || len(result.Items) != 12 {
		t.Fatalf("result=%+v maximum=%d err=%v", result, maximum.Load(), err)
	}
	if result.Items[8].ErrorCode != "QUERY_FAILED" || len(result.Items[8].Body) != 0 {
		t.Fatalf("failed result leaked detail: %+v", result.Items[8])
	}
}

func TestExecuteRejectsDuplicateSourceAndOversizedBatch(t *testing.T) {
	duplicate := fixtureItem(1)
	if _, err := Execute(context.Background(), Request{Items: []Item{duplicate, duplicate}},
		runnerFunc(successfulRun)); !errors.Is(err, ErrInvalid) {
		t.Fatal(err)
	}
	items := make([]Item, MaximumQueries+1)
	for index := range items {
		items[index] = fixtureItem(index + 1)
	}
	if _, err := Execute(context.Background(), Request{Items: items},
		runnerFunc(successfulRun)); !errors.Is(err, ErrInvalid) {
		t.Fatal(err)
	}
}

type runnerFunc func(context.Context, Item) ([]byte, error)

func (run runnerFunc) Run(
	ctx context.Context, item Item,
) ([]byte, error) {
	return run(ctx, item)
}

func successfulRun(context.Context, Item) ([]byte, error) {
	return []byte(`{"items":[]}`), nil
}

func fixtureItem(index int) Item {
	return Item{
		DataSourceID: "source_" + fmt.Sprint(index),
		QueryID:      uuid.MustParse(fmt.Sprintf("019fa000-0000-7000-8000-%012d", index)),
		VersionID:    uuid.MustParse("019fa000-0000-7000-8000-000000000101"),
		Parameters:   []byte(`{}`),
	}
}
