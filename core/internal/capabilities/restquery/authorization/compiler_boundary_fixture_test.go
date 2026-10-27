package authorization

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/adapters/postgres/compiler"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/cursor"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolved"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolver"
	"github.com/Onurlulardan/harpy/core/internal/platform/config"
	"github.com/Onurlulardan/harpy/core/internal/platform/configfixture"
)

func compilerBoundaryCodec(t *testing.T) cursor.Codec {
	t.Helper()
	key := func(value byte) string {
		return base64.RawURLEncoding.EncodeToString(bytes.Repeat([]byte{value}, 32))
	}
	keyring := fmt.Sprintf(
		`{"version":1,"current":{"kid":"current-1","key":"%s"},`+
			`"previous":{"kid":"previous-1","key":"%s"}}`, key(1), key(2),
	)
	t.Setenv("FORGECRUD_ENVIRONMENT", "development")
	t.Setenv("FORGECRUD_POSTGRES_ENDPOINT_KIND", "direct_postgres")
	t.Setenv("FORGECRUD_POSTGRES_DSN",
		"postgresql://user@example.invalid/forgecrud?sslmode=disable")
	t.Setenv("FORGECRUD_REST_QUERY_CURSOR_KEYRING", keyring)
	configfixture.SetAPI(t)
	loaded, err := config.LoadAPI()
	if err != nil {
		t.Fatal(err)
	}
	keys, err := loaded.RestQueryCursorKeyring()
	if err != nil {
		t.Fatal(err)
	}
	codec, err := cursor.NewCodec(keys)
	if err != nil {
		t.Fatal(err)
	}
	return codec
}

func compilerBoundaryCursor(
	t *testing.T, codec cursor.Codec, statement compiler.RecordStatement,
) string {
	t.Helper()
	context, err := statement.CursorContext()
	if err != nil {
		t.Fatal(err)
	}
	value, _ := resolved.NewScalarValue(
		api.LogicalTypeUUID, "aaaaaaaa-bbbb-4ccc-8ddd-eeeeeeeeeeee",
	)
	last, _ := cursor.NewLastValue(value)
	token, err := codec.Encode(context, []cursor.LastValue{last})
	if err != nil {
		t.Fatal(err)
	}
	return token
}

func compilerBoundaryAfterSpec(
	t *testing.T, spec resolved.QuerySpec, token string,
) resolved.QuerySpec {
	t.Helper()
	catalogs, _ := spec.Catalogs()
	page, _ := api.NewRequestPage(25, &token)
	request, err := api.NewRecordQueryRequest(api.RecordQueryRequestInput{
		Select: []string{"name"}, Page: page, Count: api.CountModeNone,
	})
	if err != nil {
		t.Fatal(err)
	}
	result, err := resolver.Resolve(request, catalogs)
	if err != nil {
		t.Fatal(err)
	}
	return result
}

func compilerBoundaryExactSpec(
	t *testing.T, spec resolved.QuerySpec,
) resolved.QuerySpec {
	t.Helper()
	catalogs, _ := spec.Catalogs()
	page, _ := api.NewRequestPage(25, nil)
	request, err := api.NewRecordQueryRequest(api.RecordQueryRequestInput{
		Select: []string{"name"}, Page: page, Count: api.CountModeExact,
	})
	if err != nil {
		t.Fatal(err)
	}
	result, err := resolver.Resolve(request, catalogs)
	if err != nil {
		t.Fatal(err)
	}
	return result
}

func compilerBoundaryExactAfterSpec(
	t *testing.T, spec resolved.QuerySpec, token string,
) resolved.QuerySpec {
	t.Helper()
	catalogs, _ := spec.Catalogs()
	page, _ := api.NewRequestPage(25, &token)
	request, err := api.NewRecordQueryRequest(api.RecordQueryRequestInput{
		Select: []string{"name"}, Page: page, Count: api.CountModeExact,
	})
	if err != nil {
		t.Fatal(err)
	}
	result, err := resolver.Resolve(request, catalogs)
	if err != nil {
		t.Fatal(err)
	}
	return result
}

func compilerBoundaryExactIncludeSpec(
	t *testing.T, spec resolved.QuerySpec,
) resolved.QuerySpec {
	t.Helper()
	catalogs, _ := spec.Catalogs()
	page, _ := api.NewRequestPage(25, nil)
	request, err := api.NewRecordQueryRequest(api.RecordQueryRequestInput{
		Select: []string{"name"}, Page: page, Count: api.CountModeExact,
		Include: snapshotIncludes(t, 2),
	})
	if err != nil {
		t.Fatal(err)
	}
	result, err := resolver.Resolve(request, catalogs)
	if err != nil {
		t.Fatal(err)
	}
	return result
}
