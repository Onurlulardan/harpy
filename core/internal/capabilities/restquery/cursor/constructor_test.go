package cursor

import (
	"bytes"
	"crypto/sha256"
	"strings"
	"testing"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/platform/config"
)

func TestProductionCodecRequiresValidatedKeyring(t *testing.T) {
	if _, err := NewCodec(config.CursorKeyring{}); err == nil {
		t.Fatal("zero keyring accepted")
	}
	keyring := testKeyring(t, "current", 1, "previous", 2)
	if _, err := newCodec(keyring, nil, time.Now); err == nil {
		t.Fatal("nil nonce source accepted")
	}
	if _, err := newCodec(keyring, bytes.NewReader(nil), nil); err == nil {
		t.Fatal("nil clock accepted")
	}
	codec, err := NewCodec(keyring)
	if err != nil || codec.Validate() != nil {
		t.Fatalf("validated keyring rejected: %v", err)
	}
	context := buildContext(t, baseContextFixture(t))
	token, err := codec.Encode(
		context, []LastValue{scalarLast(t, api.LogicalTypeText, "record-a")},
	)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := codec.Decode(token, context); err != nil {
		t.Fatal(err)
	}
}

func TestSortDefinitionEnforcesScalarMatrixUUIDAndBounds(t *testing.T) {
	accepted := []api.LogicalType{
		api.LogicalTypeText, api.LogicalTypeInteger, api.LogicalTypeDecimal,
		api.LogicalTypeBoolean, api.LogicalTypeDate, api.LogicalTypeTimestamp,
		api.LogicalTypeUUID, api.LogicalTypeEnum, api.LogicalTypeRelation,
	}
	for _, kind := range accepted {
		if _, err := NewSortDefinition(
			testFieldID, api.DirectionAscending, kind, maximumCanonicalValueBytes,
		); err != nil {
			t.Fatalf("sortable type %s rejected: %v", kind, err)
		}
	}
	for _, kind := range []api.LogicalType{
		api.LogicalTypeTextList, api.LogicalTypeUUIDList, api.LogicalTypeJSON,
	} {
		if _, err := NewSortDefinition(testFieldID, api.DirectionAscending, kind, 1); err == nil {
			t.Fatalf("non-scalar type %s accepted", kind)
		}
	}
	for _, test := range []struct {
		id    string
		bound int
	}{
		{"not-a-uuid", 1}, {strings.ToUpper("aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa"), 1}, {testFieldID, 0},
		{testFieldID, maximumCanonicalValueBytes + 1},
	} {
		if _, err := NewSortDefinition(
			test.id, api.DirectionAscending, api.LogicalTypeText, test.bound,
		); err == nil {
			t.Fatalf("invalid sort accepted: %+v", test)
		}
	}
}

func TestContextRejectsZeroDigestsAndMalformedMandatoryJSON(t *testing.T) {
	for _, mandatory := range [][]byte{
		nil, []byte(`{"a":1,"a":2}`), []byte(`{"a":1}{}`),
		append([]byte(`"`), append(make([]byte, maximumMandatoryBytes), '"')...),
	} {
		fixture := baseContextFixture(t)
		fixture.mandatory = mandatory
		if _, err := tryBuildContext(fixture); err == nil {
			t.Fatal("invalid mandatory predicate accepted")
		}
	}
	for _, zeroAuth := range []bool{true, false} {
		fixture := baseContextFixture(t)
		if zeroAuth {
			fixture.auth = [sha256.Size]byte{}
		} else {
			fixture.query = [sha256.Size]byte{}
		}
		if _, err := tryBuildContext(fixture); err == nil {
			t.Fatal("zero digest accepted")
		}
	}
	for _, identity := range []struct {
		tenant  string
		subject string
	}{
		{"00000000-0000-0000-0000-000000000000", testSubjectID},
		{"aaaaaaaa-aaaa-4aaa-0aaa-aaaaaaaaaaaa", testSubjectID},
		{testTenantID, "BBBBBBBB-BBBB-4BBB-8BBB-BBBBBBBBBBBB"},
	} {
		fixture := baseContextFixture(t)
		fixture.tenant, fixture.subject = identity.tenant, identity.subject
		if _, err := tryBuildContext(fixture); err == nil {
			t.Fatal("non-canonical tenant or subject accepted")
		}
	}
}
