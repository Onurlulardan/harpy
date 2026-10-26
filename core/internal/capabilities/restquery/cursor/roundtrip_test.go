package cursor

import (
	"bytes"
	"encoding/base64"
	"strings"
	"testing"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
)

func TestCodecUsesExactWireAndRoundTripsTypedLastValues(t *testing.T) {
	fixture := baseContextFixture(t)
	fixture.sorts = []SortDefinition{
		mustSort(t, testFieldID, api.DirectionAscending, api.LogicalTypeText, 32),
		mustSort(t, "66666666-6666-4666-8666-666666666666", api.DirectionDescending, api.LogicalTypeInteger, 20),
		mustSort(t, "77777777-7777-4777-8777-777777777777", api.DirectionAscending, api.LogicalTypeDecimal, 30),
		mustSort(t, "88888888-8888-4888-8888-888888888888", api.DirectionDescending, api.LogicalTypeBoolean, 5),
		mustSort(t, "99999999-9999-4999-8999-999999999999", api.DirectionAscending, api.LogicalTypeDate, 10),
		mustSort(t, "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa", api.DirectionAscending, api.LogicalTypeTimestamp, 27),
	}
	context := buildContext(t, fixture)
	values := []LastValue{
		scalarLast(t, api.LogicalTypeText, "İstanbul"),
		scalarLast(t, api.LogicalTypeInteger, "-42"),
		scalarLast(t, api.LogicalTypeDecimal, "123.450"),
		booleanLast(t, true), NewNullLastValue(),
		scalarLast(t, api.LogicalTypeTimestamp, "2026-07-15T12:00:00.000000Z"),
	}
	keyring := testKeyring(t, "current-1", 1, "previous-1", 2)
	codec, err := newCodec(keyring, bytes.NewReader(bytes.Repeat([]byte{7}, 12)), func() time.Time { return testNow })
	if err != nil {
		t.Fatal(err)
	}
	token, err := codec.Encode(context, values)
	if err != nil {
		t.Fatal(err)
	}
	segments := strings.Split(token, ".")
	if len(segments) != 4 || segments[0] != "v1" || segments[1] != "current-1" {
		t.Fatalf("wire = %q", token)
	}
	nonce, err := base64.RawURLEncoding.Strict().DecodeString(segments[2])
	if err != nil || len(nonce) != 12 || base64.RawURLEncoding.EncodeToString(nonce) != segments[2] {
		t.Fatal("nonce is not canonical 96-bit base64url")
	}
	claims, err := codec.Decode(token, context)
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := claims.LastValues()
	if err != nil || len(decoded) != len(values) {
		t.Fatalf("decoded last values = %d, %v", len(decoded), err)
	}
	if null, _ := decoded[4].IsNull(); !null {
		t.Fatal("null last value was not preserved")
	}
	issuedAt, _ := claims.IssuedAt()
	expiresAt, _ := claims.ExpiresAt()
	if !issuedAt.Equal(testNow) || expiresAt.Sub(issuedAt) != cursorLifetime {
		t.Fatal("cursor lifetime was not exactly fifteen minutes")
	}
}
