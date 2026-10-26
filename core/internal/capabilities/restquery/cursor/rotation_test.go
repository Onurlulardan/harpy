package cursor

import (
	"bytes"
	"testing"
	"time"

	restqueryv1 "github.com/Onurlulardan/harpy/contracts/restquery/v1"
)

func TestCursorRotationPreviousKeyAndMultiInstance(t *testing.T) {
	context := buildContext(t, baseContextFixture(t))
	last := []LastValue{scalarLast(t, 1, "record-a")}
	oldRing := testKeyring(t, "key-1", 1, "key-0", 9)
	oldCodec, err := newCodec(
		oldRing, bytes.NewReader(bytes.Repeat([]byte{1}, nonceBytes)),
		func() time.Time { return testNow },
	)
	if err != nil {
		t.Fatal(err)
	}
	token, err := oldCodec.Encode(context, last)
	if err != nil {
		t.Fatal(err)
	}
	rotatedRing := testKeyring(t, "key-2", 2, "key-1", 1)
	first, err := newCodec(rotatedRing, bytes.NewReader(nil), func() time.Time { return testNow })
	if err != nil {
		t.Fatal(err)
	}
	second, err := newCodec(rotatedRing, bytes.NewReader(nil), func() time.Time { return testNow })
	if err != nil {
		t.Fatal(err)
	}
	if _, err := first.Decode(token, context); err != nil {
		t.Fatalf("rotated instance rejected previous key: %v", err)
	}
	if _, err := second.Decode(token, context); err != nil {
		t.Fatalf("second instance rejected shared keyring cursor: %v", err)
	}
}

func TestEmergencyRevocationAndUnknownKIDFailClosed(t *testing.T) {
	context := buildContext(t, baseContextFixture(t))
	last := []LastValue{scalarLast(t, 1, "record-a")}
	oldRing := testKeyring(t, "key-1", 1, "key-0", 9)
	oldCodec, _ := newCodec(
		oldRing, bytes.NewReader(bytes.Repeat([]byte{1}, nonceBytes)),
		func() time.Time { return testNow },
	)
	token, err := oldCodec.Encode(context, last)
	if err != nil {
		t.Fatal(err)
	}
	revokedRing := testKeyring(t, "key-2", 2, "emergency", 3)
	revokedCodec, _ := newCodec(
		revokedRing, bytes.NewReader(nil), func() time.Time { return testNow },
	)
	_, err = revokedCodec.Decode(token, context)
	requireQueryCode(t, err, restqueryv1.ErrorQueryCursorInvalid)

	segments := splitToken(t, token)
	segments[1] = "unknown"
	_, err = oldCodec.Decode(joinToken(segments), context)
	requireQueryCode(t, err, restqueryv1.ErrorQueryCursorInvalid)
}
