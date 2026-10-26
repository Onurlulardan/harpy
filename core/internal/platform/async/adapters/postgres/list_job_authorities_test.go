package postgres

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestListJobAuthoritiesUsesOpaqueBoundedPageToken(t *testing.T) {
	t.Parallel()
	first := newWorkerTestRepository()
	second := newWorkerTestRepository()
	for _, limit := range []int{0, 257} {
		if _, err := first.BeginJobAuthorityPage(limit); !errors.Is(
			err, ErrJobAuthorityPageLimit,
		) {
			t.Fatalf("page limit %d error = %v", limit, err)
		}
	}
	for _, limit := range []int{1, 256} {
		token, err := first.BeginJobAuthorityPage(limit)
		if err != nil || !token.valid(first) {
			t.Fatalf("page limit %d token = %#v, %v", limit, token, err)
		}
	}
	token, err := first.BeginJobAuthorityPage(2)
	if err != nil || !token.valid(first) {
		t.Fatalf("first page token = %#v, %v", token, err)
	}
	if _, err := second.ListJobAuthorities(context.Background(), token); !errors.Is(
		err, ErrJobAuthorityToken,
	) {
		t.Fatalf("foreign page token error = %v", err)
	}
}

func TestListJobAuthoritiesBuildsKeysetCursorWithoutGenerationInput(t *testing.T) {
	t.Parallel()
	repository := newWorkerTestRepository()
	token, err := repository.BeginJobAuthorityPage(2)
	if err != nil {
		t.Fatal(err)
	}
	authorities := make([]JobAuthority, 3)
	for index := range authorities {
		authorities[index], err = newJobAuthority(
			repository.identity,
			"tenant",
			uuid.New(),
			"forms",
			int64(index+1),
			"forms",
			int64(index+10),
		)
		if err != nil {
			t.Fatal(err)
		}
	}
	page := makeJobAuthorityPage(token, authorities)
	values := page.Authorities()
	if len(values) != 2 {
		t.Fatalf("page size = %d", len(values))
	}
	values[0] = JobAuthority{}
	if page.Authorities()[0].scopeID == uuid.Nil {
		t.Fatal("page exposed its authority slice")
	}
	next, ok := page.Next()
	if !ok || !next.valid(repository) ||
		next.cursor.scopeID != authorities[1].scopeID {
		t.Fatalf("next cursor = %#v, %t", next, ok)
	}
	cursorType := reflect.TypeOf(next.cursor)
	if _, present := cursorType.FieldByName("groupGeneration"); present {
		t.Fatal("pagination cursor exposed a generation field")
	}
	if _, present := cursorType.FieldByName("generation"); present {
		t.Fatal("pagination cursor exposed a cutover generation field")
	}
}

func TestListJobAuthoritiesReadsCurrentV2AuthorityPage(t *testing.T) {
	t.Parallel()
	repository := newWorkerTestRepository()
	firstID, secondID, thirdID := uuid.New(), uuid.New(), uuid.New()
	database := &workerTestDatabase{queryRows: &workerTestRows{values: [][]any{
		{"tenant", firstID, "forms", int64(1), "forms", int64(11), "v2"},
		{"tenant", secondID, "workflow", int64(2), "workflow", int64(12), "v2"},
		{"tenant", thirdID, "documents", int64(3), "documents", int64(13),
			"cutover_shadow"},
	}}}
	repository.database = database
	token, err := repository.BeginJobAuthorityPage(2)
	if err != nil {
		t.Fatal(err)
	}
	page, err := repository.ListJobAuthorities(context.Background(), token)
	if err != nil {
		t.Fatal(err)
	}
	authorities := page.Authorities()
	next, hasNext := page.Next()
	if len(authorities) != 2 || authorities[0].scopeID != firstID ||
		authorities[0].mode != jobAuthorityModeV2 ||
		authorities[1].generation != 12 || !hasNext ||
		next.cursor.scopeID != secondID || database.querySQL != listJobAuthoritiesSQL {
		t.Fatalf("authority page = %#v, next %#v", authorities, next)
	}
	if len(database.queryArgs) != 6 || database.queryArgs[0] != false ||
		database.queryArgs[5] != 3 {
		t.Fatalf("authority query arguments = %#v", database.queryArgs)
	}
}

func newWorkerTestRepository() *workerRepository {
	return &workerRepository{
		database: &workerTestDatabase{}, cleanup: time.Second,
		identity: &workerIdentity{marker: 1},
	}
}
