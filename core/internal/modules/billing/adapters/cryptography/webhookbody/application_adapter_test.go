package webhookbody

import (
	"bytes"
	"context"
	"testing"
)

func TestApplicationAdaptersPreserveEncryptedEnvelopeBoundary(t *testing.T) {
	sealer, err := NewSealer(
		bytes.NewReader(bytes.Repeat([]byte{0x24}, dekSize+nonceSize)),
		testKeyWrapper{},
	)
	if err != nil {
		t.Fatal(err)
	}
	protector, err := NewApplicationProtector(sealer)
	if err != nil {
		t.Fatal(err)
	}
	reconciliationOpener, _ := NewReconciliationOpener(testKeyWrapper{})
	opener, err := NewApplicationOpener(reconciliationOpener)
	if err != nil {
		t.Fatal(err)
	}
	aad := []byte("billing-webhook-envelope.v1")
	body := []byte(`{"id":"evt_100"}`)
	protected, err := protector.Protect(context.Background(), aad, body)
	if err != nil || bytes.Equal(protected.Ciphertext(), body) {
		t.Fatalf("protected body error=%v", err)
	}
	opened, err := opener.OpenForReconciliation(context.Background(), aad, protected)
	if err != nil || !bytes.Equal(opened, body) {
		t.Fatalf("opened=%q error=%v", opened, err)
	}
	if _, err := opener.OpenForReconciliation(
		context.Background(), []byte("wrong"), protected,
	); err == nil {
		t.Fatal("wrong webhook identity opened encrypted body")
	}
}
