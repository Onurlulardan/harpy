package billingv1

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/httpapi"
)

func TestCollectCatalogPagesExhaustsCanonicalCursors(t *testing.T) {
	t.Parallel()
	handler := &Handler{query: billingapi.NewQueryDecoder()}
	cursors := []string{"cursor-one", "cursor-two"}
	call := 0
	items, err := handler.collectCatalogPages(context.Background(), func(
		_ context.Context, request billingapi.QueryRequest,
	) (billingapi.QueryPage, error) {
		page := request.Page()
		if page.Size() != 500 {
			t.Fatalf("page size = %d", page.Size())
		}
		after, hasAfter := page.After()
		if call == 0 && hasAfter || call > 0 && (!hasAfter || after != cursors[call-1]) {
			t.Fatalf("call %d cursor = %q, %t", call, after, hasAfter)
		}
		result := billingapi.QueryPage{Items: []json.RawMessage{json.RawMessage(`{}`)}}
		if call < len(cursors) {
			result.Pagination.Next = &cursors[call]
		}
		call++
		return result, nil
	})
	if err != nil || call != 3 || len(items) != 3 {
		t.Fatalf("pages=%d items=%d error=%v", call, len(items), err)
	}
}

func TestCollectCatalogPagesRejectsCursorCycle(t *testing.T) {
	t.Parallel()
	handler := &Handler{query: billingapi.NewQueryDecoder()}
	cursor := "same-cursor"
	_, err := handler.collectCatalogPages(context.Background(), func(
		context.Context, billingapi.QueryRequest,
	) (billingapi.QueryPage, error) {
		return billingapi.QueryPage{Pagination: billingapi.Pagination{Next: &cursor}}, nil
	})
	if !errors.Is(err, ErrPaginationEvidence) {
		t.Fatalf("cursor cycle error = %v", err)
	}
}

func TestCollectCatalogPagesFailsClosedAtHardBudget(t *testing.T) {
	t.Parallel()
	handler := &Handler{query: billingapi.NewQueryDecoder()}
	call := 0
	_, err := handler.collectCatalogPages(context.Background(), func(
		context.Context, billingapi.QueryRequest,
	) (billingapi.QueryPage, error) {
		call++
		next := fmt.Sprintf("cursor-%03d", call)
		return billingapi.QueryPage{Pagination: billingapi.Pagination{Next: &next}}, nil
	})
	if call != maximumPageCount || !errors.Is(err, ErrPaginationEvidence) {
		t.Fatalf("page budget calls=%d error=%v", call, err)
	}
}
