package cursor

import (
	"bytes"
	"crypto/sha256"
	"strings"
	"testing"
	"time"

	restqueryv1 "github.com/Onurlulardan/harpy/contracts/restquery/v1"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
)

func TestDecodeRejectsEverySecurityContextMismatch(t *testing.T) {
	fixture := baseContextFixture(t)
	context := buildContext(t, fixture)
	codec, _ := newCodec(
		testKeyring(t, "current", 1, "previous", 2),
		bytes.NewReader(bytes.Repeat([]byte{8}, nonceBytes)),
		func() time.Time { return testNow },
	)
	token, err := codec.Encode(context, []LastValue{scalarLast(t, api.LogicalTypeText, "record-a")})
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name   string
		mutate func(*contextFixture)
	}{
		{"tenant", func(value *contextFixture) { value.tenant = "aaaaaaaa-1111-4111-8111-111111111111" }},
		{"subject", func(value *contextFixture) { value.subject = "bbbbbbbb-2222-4222-8222-222222222222" }},
		{"auth", func(value *contextFixture) { value.auth = sha256.Sum256([]byte("other-auth")) }},
		{"query", func(value *contextFixture) { value.query = sha256.Sum256([]byte("other-query")) }},
		{"mandatory", func(value *contextFixture) { value.mandatory = []byte(`{"active":false}`) }},
		{"page", func(value *contextFixture) { value.pageSize++ }},
		{"sort field", func(value *contextFixture) {
			value.sorts = []SortDefinition{mustSort(
				t, "cccccccc-5555-4555-8555-555555555555",
				api.DirectionAscending, api.LogicalTypeText, 128,
			)}
		}},
		{"sort direction", func(value *contextFixture) {
			value.sorts = []SortDefinition{mustSort(
				t, testFieldID, api.DirectionDescending, api.LogicalTypeText, 128,
			)}
		}},
		{"sort canonical bound", func(value *contextFixture) {
			value.sorts = []SortDefinition{mustSort(
				t, testFieldID, api.DirectionAscending, api.LogicalTypeText, 1,
			)}
		}},
		{"other owner", func(value *contextFixture) { value.owner = "crm" }},
		{"other resource", func(value *contextFixture) { value.resource = "dddddddd-3333-4333-8333-333333333333" }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			changed := baseContextFixture(t)
			test.mutate(&changed)
			_, decodeErr := codec.Decode(token, buildContext(t, changed))
			requireQueryCode(t, decodeErr, restqueryv1.ErrorQueryCursorInvalid)
		})
	}
}

func TestDecodeTreatsPublishedVersionAndCatalogHashAsStale(t *testing.T) {
	context := buildContext(t, baseContextFixture(t))
	codec, _ := newCodec(
		testKeyring(t, "current", 1, "previous", 2),
		bytes.NewReader(bytes.Repeat([]byte{9}, nonceBytes)),
		func() time.Time { return testNow },
	)
	token, err := codec.Encode(context, []LastValue{scalarLast(t, api.LogicalTypeText, "record-a")})
	if err != nil {
		t.Fatal(err)
	}
	for _, changed := range []contextFixture{
		func() contextFixture {
			value := baseContextFixture(t)
			value.version = "eeeeeeee-4444-4444-8444-444444444444"
			return value
		}(),
		func() contextFixture {
			value := baseContextFixture(t)
			value.hash = strings.Repeat("b", sha256.Size*2)
			return value
		}(),
	} {
		_, decodeErr := codec.Decode(token, buildContext(t, changed))
		requireQueryCode(t, decodeErr, restqueryv1.ErrorQueryCursorStale)
	}
}

func TestEquivalentCanonicalMandatoryPredicateBinds(t *testing.T) {
	context := buildContext(t, baseContextFixture(t))
	codec, _ := newCodec(
		testKeyring(t, "current", 1, "previous", 2),
		bytes.NewReader(bytes.Repeat([]byte{10}, nonceBytes)),
		func() time.Time { return testNow },
	)
	token, err := codec.Encode(context, []LastValue{scalarLast(t, api.LogicalTypeText, "record-a")})
	if err != nil {
		t.Fatal(err)
	}
	equivalent := baseContextFixture(t)
	equivalent.mandatory = []byte(` { "active" : true, "tenant" : "` + testTenantID + `" } `)
	if _, err := codec.Decode(token, buildContext(t, equivalent)); err != nil {
		t.Fatalf("equivalent canonical predicate rejected: %v", err)
	}
}
