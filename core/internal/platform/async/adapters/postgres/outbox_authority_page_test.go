package postgres

import (
	"context"
	"errors"
	"testing"
)

func TestOutboxAuthorityPageIsStoreBoundAndBounded(t *testing.T) {
	t.Parallel()
	repository := newWorkerTestRepository()
	for _, limit := range []int{0, maximumOutboxAuthorityPageSize + 1} {
		if _, err := repository.BeginOutboxAuthorityPage(limit); !errors.Is(
			err, ErrOutboxAuthorityPageLimit,
		) {
			t.Fatalf("limit %d error = %v", limit, err)
		}
	}
	token, err := repository.BeginOutboxAuthorityPage(2)
	if err != nil {
		t.Fatal(err)
	}
	other := newWorkerTestRepository()
	if _, err := other.ListOutboxAuthorities(
		context.Background(), token,
	); !errors.Is(err, ErrOutboxAuthorityToken) {
		t.Fatalf("foreign token error = %v", err)
	}
}

func TestOutboxAuthorityPageReturnsOpaqueContinuation(t *testing.T) {
	t.Parallel()
	repository := newWorkerTestRepository()
	token, err := repository.BeginOutboxAuthorityPage(1)
	if err != nil {
		t.Fatal(err)
	}
	first := validOutboxAuthority(t, repository)
	second, err := newOutboxAuthority(
		repository.identity, "tenant", first.scopeID,
		"operations", 3, "workflow_second", 8,
	)
	if err != nil {
		t.Fatal(err)
	}
	page := makeOutboxAuthorityPage(token, []OutboxAuthority{first, second})
	if got := page.Authorities(); len(got) != 1 || got[0].domain != first.domain {
		t.Fatalf("authorities = %#v", got)
	}
	next, ok := page.Next()
	if !ok || !next.valid(repository) || !next.cursor.present ||
		next.cursor.domain != first.domain {
		t.Fatalf("next token = %#v, %v", next, ok)
	}
}
