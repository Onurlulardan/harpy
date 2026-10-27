package domain

import (
	"testing"
	"time"
)

func TestCommentRedactionProducesImmutableTombstone(t *testing.T) {
	author := mustID(t, "018f9f3e-7c16-7a21-8b42-116a9d3c9870")
	comment := Comment{ID: author, BoardID: author, TaskID: author, AuthorID: author,
		Body: "original", CreatedAt: time.Unix(1, 0).UTC(), RowVersion: 1}
	redacted, err := comment.Redact(author, false, "", time.Unix(2, 0).UTC())
	if err != nil || redacted.Body != "" || redacted.Redaction != RedactionAuthor ||
		redacted.RowVersion != 2 || comment.Body != "original" {
		t.Fatalf("redaction = %+v, %v", redacted, err)
	}
	if _, err := redacted.Redact(author, false, "", time.Unix(3, 0).UTC()); err == nil {
		t.Fatal("redacted tombstone accepted a second mutation")
	}
}
