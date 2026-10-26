package idempotency

import (
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestRestoreReceiptVerifiesStoredResponseHash(t *testing.T) {
	t.Parallel()
	response, err := NewBodyResponse(200, map[string]string{"content-type": "application/json"}, []byte("{}"))
	if err != nil {
		t.Fatal(err)
	}
	now := time.Now().UTC()
	stored := StoredReceipt{
		ReceiptID: uuid.New(), RequestSHA256: strings.Repeat("a", 64),
		AuthorizationVersion: 3, RetentionClass: InteractiveWrite,
		Status: response.Status(), HeadersJSON: response.HeadersJSON(), Body: response.Body(),
		ResponseSHA256: response.SHA256(), CompletedAt: now, RetainUntil: now.Add(24 * time.Hour),
	}
	receipt, err := RestoreReceipt(stored)
	if err != nil || receipt.Response().SHA256() != response.SHA256() {
		t.Fatalf("restore receipt = %#v, %v", receipt, err)
	}
	stored.Body = []byte("tampered")
	if _, err := RestoreReceipt(stored); err == nil {
		t.Fatal("tampered stored response accepted")
	}
}
