//go:build integration

package postgres

import (
	"context"
	"net/url"
	"strings"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/platform/contractretirement"
	"github.com/jackc/pgx/v5"
)

func connectContractRetirementRuntime(
	t *testing.T, ctx context.Context, dsn string,
) *pgx.Conn {
	t.Helper()
	parsed, err := url.Parse(dsn)
	if err != nil {
		t.Fatal(err)
	}
	parsed.User = url.UserPassword("harpy_v2_runtime", "integration-runtime-secret")
	configuration, err := pgx.ParseConfig(parsed.String())
	if err != nil {
		t.Fatal(err)
	}
	configuration.RuntimeParams["role"] = "harpy_runtime"
	configuration.RuntimeParams["search_path"] = "pg_catalog"
	connection, err := pgx.ConnectConfig(ctx, configuration)
	if err != nil {
		t.Fatal(err)
	}
	var sessionUser, currentUser string
	if err := connection.QueryRow(ctx,
		"SELECT session_user::text, current_user::text",
	).Scan(&sessionUser, &currentUser); err != nil {
		t.Fatal(err)
	}
	if sessionUser != "harpy_v2_runtime" || currentUser != "harpy_runtime" {
		t.Fatalf("runtime identity = %s/%s", sessionUser, currentUser)
	}
	return connection
}

func contractRetirementKey(t *testing.T, token string) contractretirement.Key {
	t.Helper()
	key, err := contractretirement.ParseKey("rq.alias." + strings.Repeat(token, 64))
	if err != nil {
		t.Fatal(err)
	}
	return key
}

func contractRetirementBatch(
	t *testing.T, identity contractretirement.Identity, keys ...contractretirement.Key,
) contractretirement.Batch {
	t.Helper()
	batch, err := contractretirement.NewBatch(identity, keys)
	if err != nil {
		t.Fatal(err)
	}
	return batch
}

func contractRetirementCount(
	t *testing.T, value int64,
) contractretirement.RemainingScopeCount {
	t.Helper()
	count, err := contractretirement.NewRemainingScopeCount(value)
	if err != nil {
		t.Fatal(err)
	}
	return count
}

func revisionValue(t *testing.T, revision contractretirement.Revision) int64 {
	t.Helper()
	value, err := revision.DatabaseValue()
	if err != nil {
		t.Fatal(err)
	}
	return value
}

func keyValue(t *testing.T, key contractretirement.Key) string {
	t.Helper()
	value, err := key.DatabaseValue()
	if err != nil {
		t.Fatal(err)
	}
	return value
}
