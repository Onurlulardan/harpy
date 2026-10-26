package archive

import (
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/modules/documents/domain"
)

func TestRequestDigestIsOrderIndependentAndRetryBound(t *testing.T) {
	first := digestID(t, "018f0000-0000-7000-8000-000000000121")
	second := digestID(t, "018f0000-0000-7000-8000-000000000122")
	retry := digestID(t, "018f0000-0000-7000-8000-000000000123")
	left, valid := RequestDigest(CreateRequest{ItemIDs: []domain.ID{first, second}})
	right, reordered := RequestDigest(CreateRequest{ItemIDs: []domain.ID{second, first}})
	bound, withRetry := RequestDigest(CreateRequest{ItemIDs: []domain.ID{first, second},
		RetryOfArchiveID: retry})
	if !valid || !reordered || !withRetry || left != right || left == bound {
		t.Fatalf("unexpected digest contract: valid=%v reordered=%v retry=%v",
			valid, reordered, withRetry)
	}
}

func digestID(t *testing.T, value string) domain.ID {
	t.Helper()
	id, err := domain.ParseID(value)
	if err != nil {
		t.Fatal(err)
	}
	return id
}
