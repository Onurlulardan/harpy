package cursor

import (
	"bytes"
	"strconv"
	"strings"
	"testing"
	"time"

	restqueryv1 "github.com/Onurlulardan/harpy/contracts/restquery/v1"
)

func TestCursorClockSkewAndExpiryBoundaries(t *testing.T) {
	context := buildContext(t, baseContextFixture(t))
	now := testNow
	codec, err := newCodec(
		testKeyring(t, "current", 1, "previous", 2),
		bytes.NewReader(bytes.Repeat([]byte{3}, nonceBytes)),
		func() time.Time { return now },
	)
	if err != nil {
		t.Fatal(err)
	}
	token, err := codec.Encode(context, []LastValue{scalarLast(t, 1, "record-a")})
	if err != nil {
		t.Fatal(err)
	}
	now = testNow.Add(-allowedClockSkew)
	if _, err := codec.Decode(token, context); err != nil {
		t.Fatalf("exact future skew rejected: %v", err)
	}
	now = testNow.Add(-allowedClockSkew - time.Nanosecond)
	_, err = codec.Decode(token, context)
	requireQueryCode(t, err, restqueryv1.ErrorQueryCursorInvalid)
	now = testNow.Add(cursorLifetime + allowedClockSkew)
	if _, err := codec.Decode(token, context); err != nil {
		t.Fatalf("exact expiry skew rejected: %v", err)
	}
	now = testNow.Add(cursorLifetime + allowedClockSkew + time.Nanosecond)
	_, err = codec.Decode(token, context)
	requireQueryCode(t, err, restqueryv1.ErrorQueryCursorExpired)
}

func TestClaimLifetimeMustBeExactlyFifteenMinutes(t *testing.T) {
	codec, context, token := encodedFixture(t)
	plaintext, key, err := codec.decryptToken(token)
	if err != nil {
		t.Fatal(err)
	}
	defer clear(plaintext)
	expires := testNow.Add(cursorLifetime).Unix()
	tampered := bytes.Replace(
		plaintext, []byte(strconv.FormatInt(expires, 10)),
		[]byte(strconv.FormatInt(expires+1, 10)), 1,
	)
	if bytes.Equal(tampered, plaintext) {
		t.Fatal("expiry mutation did not change plaintext")
	}
	_, err = codec.Decode(resealPlaintext(t, key, tampered, 30), context)
	requireQueryCode(t, err, restqueryv1.ErrorQueryCursorInvalid)
}

func TestExpiredWinsBeforeStaleCatalogClassification(t *testing.T) {
	context := buildContext(t, baseContextFixture(t))
	now := testNow
	codec, _ := newCodec(
		testKeyring(t, "current", 1, "previous", 2),
		bytes.NewReader(bytes.Repeat([]byte{4}, nonceBytes)),
		func() time.Time { return now },
	)
	token, err := codec.Encode(context, []LastValue{scalarLast(t, 1, "record-a")})
	if err != nil {
		t.Fatal(err)
	}
	staleFixture := baseContextFixture(t)
	staleFixture.hash = strings.Repeat("b", 64)
	now = testNow.Add(cursorLifetime + allowedClockSkew + time.Second)
	_, err = codec.Decode(token, buildContext(t, staleFixture))
	requireQueryCode(t, err, restqueryv1.ErrorQueryCursorExpired)
}
