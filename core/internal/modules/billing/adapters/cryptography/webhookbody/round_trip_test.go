package webhookbody

import (
	"bytes"
	"context"
	"testing"
)

type testKeyWrapper struct{}

func (testKeyWrapper) Wrap(_ context.Context, dek []byte) (WrappedDEK, error) {
	wrapped := append([]byte(nil), dek...)
	for index := range wrapped {
		wrapped[index] ^= 0xa5
	}
	return WrappedDEK{KeyID: "test-kek-v1", Ciphertext: wrapped}, nil
}

func (testKeyWrapper) Unwrap(_ context.Context, wrapped WrappedDEK) ([]byte, error) {
	dek := append([]byte(nil), wrapped.Ciphertext...)
	for index := range dek {
		dek[index] ^= 0xa5
	}
	return dek, nil
}

func TestSealAndReconciliationOpen(t *testing.T) {
	random := bytes.NewReader(bytes.Repeat([]byte{0x42}, dekSize+nonceSize))
	sealer, err := NewSealer(random, testKeyWrapper{})
	if err != nil {
		t.Fatal(err)
	}
	opener, err := NewReconciliationOpener(testKeyWrapper{})
	if err != nil {
		t.Fatal(err)
	}
	aad := []byte("stripe/account/event-1")
	body := []byte(`{"id":"event-1","object":"subscription"}`)
	envelope, err := sealer.Seal(context.Background(), aad, body)
	if err != nil || string(envelope.Ciphertext) == string(body) {
		t.Fatalf("seal failed: %v", err)
	}
	opened, err := opener.Open(context.Background(), aad, envelope)
	if err != nil || !bytes.Equal(opened, body) {
		t.Fatalf("open=%q, err=%v", opened, err)
	}
}

func TestOpenRejectsTamperAndWrongAssociatedData(t *testing.T) {
	for _, mutate := range []func(*Envelope){
		func(envelope *Envelope) { envelope.Ciphertext[0] ^= 1 },
		func(envelope *Envelope) { envelope.BodySHA256[0] ^= 1 },
	} {
		envelope := sealedFixture(t)
		mutate(&envelope)
		opener, _ := NewReconciliationOpener(testKeyWrapper{})
		if _, err := opener.Open(context.Background(), []byte("provider/account/event"), envelope); err == nil {
			t.Fatal("tampered envelope was accepted")
		}
	}
	envelope := sealedFixture(t)
	opener, _ := NewReconciliationOpener(testKeyWrapper{})
	if _, err := opener.Open(context.Background(), []byte("wrong"), envelope); err == nil {
		t.Fatal("wrong associated data was accepted")
	}
}

func TestSealEnforcesBodyBudget(t *testing.T) {
	sealer, _ := NewSealer(
		bytes.NewReader(bytes.Repeat([]byte{1}, dekSize+nonceSize)), testKeyWrapper{})
	for _, body := range [][]byte{nil, make([]byte, MaximumBodySize+1)} {
		if _, err := sealer.Seal(context.Background(), []byte("aad"), body); err != ErrBody {
			t.Fatalf("expected ErrBody, got %v", err)
		}
	}
}

func sealedFixture(t *testing.T) Envelope {
	t.Helper()
	sealer, _ := NewSealer(
		bytes.NewReader(bytes.Repeat([]byte{7}, dekSize+nonceSize)), testKeyWrapper{})
	envelope, err := sealer.Seal(
		context.Background(), []byte("provider/account/event"), []byte("body"))
	if err != nil {
		t.Fatal(err)
	}
	return envelope
}
