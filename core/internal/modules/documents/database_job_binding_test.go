package documents

import (
	"testing"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/documents/application/foundation"
)

type bindingIDs struct{ next byte }

func (source *bindingIDs) NewUUIDv7() (foundation.ID, error) {
	source.next++
	value := mustDocumentID("018f0000-0000-7000-8000-000000000700")
	value[15] = source.next
	return value, nil
}

type bindingClock struct{}

func (bindingClock) Now() time.Time {
	return time.Date(2026, 7, 18, 15, 0, 0, 0, time.UTC)
}

func TestDocumentsDatabaseJobBindingsAreRepeatableRead(t *testing.T) {
	var ids foundation.IDSource = &bindingIDs{}
	var clock foundation.Clock = bindingClock{}
	purge, err := NewPurgeJobRuntime(ids, clock)
	if err != nil || purge.Binding().Owner() != "documents" ||
		!purge.Binding().RepeatableRead() {
		t.Fatalf("purge binding=%+v err=%v", purge.Binding(), err)
	}
	expiry, err := NewArchiveExpiryJobRuntime(ids, clock)
	if err != nil || expiry.Binding().Owner() != "documents" ||
		!expiry.Binding().RepeatableRead() {
		t.Fatalf("expiry binding=%+v err=%v", expiry.Binding(), err)
	}
}
