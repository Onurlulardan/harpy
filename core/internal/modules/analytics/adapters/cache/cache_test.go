package cache

import (
	"testing"
	"time"

	queryapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/query"
	"github.com/google/uuid"
)

type mutableClock struct{ now time.Time }

func (clock *mutableClock) Now() time.Time { return clock.now }

func TestCacheKeysEveryAuthorityInputAndInvalidatesQuery(t *testing.T) {
	clock := &mutableClock{now: time.Date(2026, 7, 18, 12, 0, 0, 0, time.UTC)}
	cache, err := New(clock, 2)
	if err != nil {
		t.Fatal(err)
	}
	request := cacheRequest()
	body := []byte(`{"items":[1]}`)
	if err = cache.Write(t.Context(), request, body, clock.now.Add(time.Minute)); err != nil {
		t.Fatal(err)
	}
	body[10] = '9'
	read, hit, err := cache.Read(t.Context(), request)
	if err != nil || !hit || string(read) != `{"items":[1]}` {
		t.Fatalf("read = %s, hit = %v, err = %v", read, hit, err)
	}
	stale := request
	stale.Binding.RevisionsSHA256 = [32]byte{9}
	if _, hit, _ = cache.Read(t.Context(), stale); hit {
		t.Fatal("stale revision reused cached result")
	}
	if err = cache.Invalidate(t.Context(), request.QueryID); err != nil {
		t.Fatal(err)
	}
	if _, hit, _ = cache.Read(t.Context(), request); hit {
		t.Fatal("invalidated query remained cached")
	}
}

func TestCacheExpiresAndEvictsWithoutChangingCorrectness(t *testing.T) {
	clock := &mutableClock{now: time.Date(2026, 7, 18, 12, 0, 0, 0, time.UTC)}
	cache, _ := New(clock, 1)
	first := cacheRequest()
	_ = cache.Write(t.Context(), first, []byte(`{"v":1}`), clock.now.Add(time.Second))
	clock.now = clock.now.Add(2 * time.Second)
	if _, hit, _ := cache.Read(t.Context(), first); hit {
		t.Fatal("expired entry was returned")
	}
}

func cacheRequest() queryapp.CacheRequest {
	return queryapp.CacheRequest{TenantID: uuid.New(), QueryID: uuid.New(),
		QueryVersionID: uuid.New(), Binding: queryapp.ExecutionBinding{ReleaseGeneration: 1,
			AuthorizationSHA256: [32]byte{1}, PredicateSHA256: [32]byte{2},
			RevisionsSHA256: [32]byte{3}}, QuerySHA256: [32]byte{4},
		ParametersSHA256: [32]byte{5}}
}
