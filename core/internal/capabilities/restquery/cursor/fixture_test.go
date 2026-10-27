package cursor

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolved"
	"github.com/Onurlulardan/harpy/core/internal/platform/config"
	"github.com/Onurlulardan/harpy/core/internal/platform/configfixture"
)

const (
	testTenantID   = "11111111-1111-4111-8111-111111111111"
	testSubjectID  = "22222222-2222-4222-8222-222222222222"
	testResourceID = "33333333-3333-4333-8333-333333333333"
	testVersionID  = "44444444-4444-4444-8444-444444444444"
	testFieldID    = "55555555-5555-4555-8555-555555555555"
)

var testNow = time.Date(2026, 7, 15, 12, 0, 0, 0, time.UTC)

type contextFixture struct {
	tenant, subject string
	auth, query     [sha256.Size]byte
	owner           string
	resource        string
	version         string
	hash            string
	mandatory       []byte
	pageSize        int
	sorts           []SortDefinition
}

func baseContextFixture(t testing.TB) contextFixture {
	t.Helper()
	return contextFixture{
		tenant: testTenantID, subject: testSubjectID,
		auth:  sha256.Sum256([]byte("authorization-context")),
		query: sha256.Sum256([]byte("canonical-query-without-cursor")),
		owner: "forms", resource: testResourceID, version: testVersionID,
		hash:      strings.Repeat("a", sha256.Size*2),
		mandatory: []byte(`{"tenant":"` + testTenantID + `","active":true}`),
		pageSize:  100,
		sorts: []SortDefinition{
			mustSort(t, testFieldID, api.DirectionAscending, api.LogicalTypeText, 128),
		},
	}
}

func buildContext(t testing.TB, fixture contextFixture) Context {
	t.Helper()
	hash, err := api.NewCatalogHash(fixture.hash)
	if err != nil {
		t.Fatal(err)
	}
	reference, err := api.NewCatalogReference(
		fixture.owner, fixture.resource, fixture.version, hash,
	)
	if err != nil {
		t.Fatal(err)
	}
	context, err := NewContext(
		fixture.tenant, fixture.subject, fixture.auth, reference, fixture.query,
		fixture.mandatory, fixture.pageSize, fixture.sorts,
	)
	if err != nil {
		t.Fatal(err)
	}
	return context
}

func mustSort(
	t testing.TB, field string, direction api.Direction,
	kind api.LogicalType, bound int,
) SortDefinition {
	t.Helper()
	sort, err := NewSortDefinition(field, direction, kind, bound)
	if err != nil {
		t.Fatal(err)
	}
	return sort
}

func scalarLast(t testing.TB, kind api.LogicalType, value string) LastValue {
	t.Helper()
	resolvedValue, err := resolved.NewScalarValue(kind, value)
	if err != nil {
		t.Fatal(err)
	}
	last, err := NewLastValue(resolvedValue)
	if err != nil {
		t.Fatal(err)
	}
	return last
}

func booleanLast(t testing.TB, value bool) LastValue {
	t.Helper()
	resolvedValue, err := resolved.NewBooleanValue(value)
	if err != nil {
		t.Fatal(err)
	}
	last, err := NewLastValue(resolvedValue)
	if err != nil {
		t.Fatal(err)
	}
	return last
}

func testKeyring(
	t testing.TB, currentKID string, currentByte byte,
	previousKID string, previousByte byte,
) config.CursorKeyring {
	t.Helper()
	key := func(value byte) string {
		return base64.RawURLEncoding.EncodeToString(bytes.Repeat([]byte{value}, 32))
	}
	document := fmt.Sprintf(
		`{"version":1,"current":{"kid":%q,"key":%q},"previous":{"kid":%q,"key":%q}}`,
		currentKID, key(currentByte), previousKID, key(previousByte),
	)
	t.Setenv("FORGECRUD_ENVIRONMENT", "development")
	t.Setenv("FORGECRUD_POSTGRES_ENDPOINT_KIND", "direct_postgres")
	t.Setenv("FORGECRUD_POSTGRES_DSN", "postgresql://user@example.invalid/forgecrud?sslmode=disable")
	t.Setenv("FORGECRUD_REST_QUERY_CURSOR_KEYRING", document)
	configfixture.SetAPI(t)
	loaded, err := config.LoadAPI()
	if err != nil {
		t.Fatal(err)
	}
	keyring, err := loaded.RestQueryCursorKeyring()
	if err != nil {
		t.Fatal(err)
	}
	return keyring
}
