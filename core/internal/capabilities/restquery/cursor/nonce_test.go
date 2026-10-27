package cursor

import (
	"bytes"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
)

type failingNonceReader struct{}

func (failingNonceReader) Read([]byte) (int, error) {
	return 0, errors.New("nonce source failed")
}

func TestInjectedNonceSourceUsesExactDistinct96BitValues(t *testing.T) {
	context := buildContext(t, baseContextFixture(t))
	stream := append(
		bytes.Repeat([]byte{1}, nonceBytes), bytes.Repeat([]byte{2}, nonceBytes)...,
	)
	codec, err := newCodec(
		testKeyring(t, "current", 1, "previous", 2), bytes.NewReader(stream),
		func() time.Time { return testNow },
	)
	if err != nil {
		t.Fatal(err)
	}
	values := []LastValue{scalarLast(t, api.LogicalTypeText, "record-a")}
	first, err := codec.Encode(context, values)
	if err != nil {
		t.Fatal(err)
	}
	second, err := codec.Encode(context, values)
	if err != nil {
		t.Fatal(err)
	}
	firstNonce := strings.Split(first, ".")[2]
	secondNonce := strings.Split(second, ".")[2]
	if firstNonce == secondNonce {
		t.Fatal("deterministic distinct nonce input was reused")
	}
	if _, err := codec.Encode(context, values); err == nil {
		t.Fatal("exhausted nonce source accepted")
	}
}

func TestNonceSourceFailureIsRedacted(t *testing.T) {
	context := buildContext(t, baseContextFixture(t))
	codec, err := newCodec(
		testKeyring(t, "current", 1, "previous", 2), failingNonceReader{},
		func() time.Time { return testNow },
	)
	if err != nil {
		t.Fatal(err)
	}
	_, err = codec.Encode(
		context, []LastValue{scalarLast(t, api.LogicalTypeText, "record-a")},
	)
	if err == nil || strings.Contains(err.Error(), "nonce source failed") {
		t.Fatalf("nonce failure leaked dependency error: %v", err)
	}
}
