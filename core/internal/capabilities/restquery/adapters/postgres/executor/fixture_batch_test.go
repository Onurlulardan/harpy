package executor

import (
	"bytes"
	"crypto/sha256"
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
	"github.com/google/uuid"
)

var (
	fixtureTenant = uuid.MustParse("30000000-0000-4000-8000-000000000001")
	fixtureActor  = uuid.MustParse("30000000-0000-4000-8000-000000000002")
)

type batchFixture struct {
	batch compiler.RecordBatch
	plan  resolved.AuthorizedPlan
	codec cursor.Codec
}

func newBatchFixture(
	t *testing.T,
	count api.CountMode,
	tenant uuid.UUID,
	actor uuid.UUID,
) batchFixture {
	return newBatchFixtureWithPageSize(t, count, tenant, actor, 2)
}

func newBatchFixtureWithPageSize(
	t *testing.T,
	count api.CountMode,
	tenant uuid.UUID,
	actor uuid.UUID,
	pageSize int,
) batchFixture {
	t.Helper()
	catalogs := fixtureCatalogSet(t, "forms")
	page, _ := api.NewRequestPage(pageSize, nil)
	request, err := api.NewRecordQueryRequest(api.RecordQueryRequestInput{
		Select: []string{"name", "metadata"}, Page: page, Count: count,
	})
	if err != nil {
		t.Fatal(err)
	}
	spec, err := resolver.Resolve(request, catalogs)
	if err != nil {
		t.Fatal(err)
	}
	reference, _ := spec.Reference()
	path, _ := spec.RootPath()
	visibility, err := resolved.NewAllResourceVisibility(
		sha256.Sum256([]byte("executor-requirement")), reference, path,
	)
	if err != nil {
		t.Fatal(err)
	}
	plan, err := resolved.MintAuthorizedPlan(
		spec, tenant, actor, sha256.Sum256([]byte("executor-authorization")),
		[]resolved.ResourceVisibility{visibility},
	)
	if err != nil {
		t.Fatal(err)
	}
	codec := fixtureCodec(t)
	postgresCompiler, err := compiler.New(codec)
	if err != nil {
		t.Fatal(err)
	}
	batch, err := postgresCompiler.CompileRecordBatch(plan)
	if err != nil {
		t.Fatal(err)
	}
	return batchFixture{batch: batch, plan: plan, codec: codec}
}

func fixtureCodec(t *testing.T) cursor.Codec {
	t.Helper()
	key := func(value byte) string {
		return base64.RawURLEncoding.EncodeToString(bytes.Repeat([]byte{value}, 32))
	}
	document := fmt.Sprintf(
		`{"version":1,"current":{"kid":"current-1","key":"%s"},`+
			`"previous":{"kid":"previous-1","key":"%s"}}`, key(1), key(2),
	)
	t.Setenv("FORGECRUD_ENVIRONMENT", "development")
	t.Setenv("FORGECRUD_POSTGRES_ENDPOINT_KIND", "direct_postgres")
	t.Setenv("FORGECRUD_POSTGRES_DSN",
		"postgresql://user@example.invalid/forgecrud?sslmode=disable")
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
	codec, err := cursor.NewCodec(keyring)
	if err != nil {
		t.Fatal(err)
	}
	return codec
}
