//go:build integration

package integrationtest

import (
	"context"
	"crypto/sha256"
	"testing"

	profileapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/profile"
	profilebackfillapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/profilebackfill"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

const (
	recoveryUserV5ID     = "41000000-0000-4000-8000-000000000005"
	recoveryFutureV6ID   = "41000000-0000-4000-8000-000000000006"
	recoveryFutureUserID = "51000000-0000-4000-8000-000000000006"
)

func recoveryUserTargetV5(t *testing.T) profileapp.Schema {
	t.Helper()
	return recoverySchema(
		t, recoveryUserV5ID, profileapp.ResourceUser, 5,
		recoveryUserTargetV4(t).Fields(),
	)
}

func recoveryMappingsV5() []profilebackfillapp.Mapping {
	return []profilebackfillapp.Mapping{
		{SourceFieldID: uuid.MustParse(recoveryReplacementV4ID),
			TargetFieldID: uuid.MustParse(recoveryReplacementV4ID),
			ConverterKind: profilebackfillapp.MappingCopy},
		{SourceFieldID: uuid.MustParse(publicationOptionalID),
			TargetFieldID: uuid.MustParse(publicationOptionalID),
			ConverterKind: profilebackfillapp.MappingCopy},
	}
}

func seedRecoveryFutureUser(
	t *testing.T, ctx context.Context, db publicationDatabase,
) {
	t.Helper()
	definition := sha256.Sum256([]byte("unknown-future-schema-v6"))
	document := []byte(`{}`)
	digest := sha256.Sum256(document)
	const statement = `INSERT INTO iam_profile_schema_versions(
id,resource_kind,version,state,definition_sha256,activated_at,retired_at,
created_at,updated_at,revision) VALUES($1,'user',6,'draft',$2,NULL,NULL,
pg_catalog.now(),pg_catalog.now(),1);
INSERT INTO iam_users(
id,identity_kind,status,profile_resource_kind,profile_schema_version_id,
profile_values,profile_sha256,token_version,authorization_version,revision,
created_at,updated_at,tombstoned_at) VALUES
($3,'human','active','user',$1,$4::jsonb,$5,1,0,1,
pg_catalog.now(),pg_catalog.now(),NULL);`
	tx, err := db.admin.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	// This fixture represents corrupt future-version data that guarded writers
	// cannot create once the parent is fenced.
	_, err = tx.Exec(ctx, "SELECT pg_catalog.set_config('search_path',$1,true)",
		db.tenant.schema+",platform,pg_catalog")
	if err == nil {
		_, err = tx.Exec(ctx, `ALTER TABLE iam_users DISABLE TRIGGER
iam_users_profile_backfill_insert_guard`)
	}
	if err == nil {
		_, err = tx.Exec(ctx, statement, pgx.QueryExecModeSimpleProtocol,
			recoveryFutureV6ID, definition[:], recoveryFutureUserID,
			string(document), digest[:])
	}
	if err == nil {
		_, err = tx.Exec(ctx, `ALTER TABLE iam_users ENABLE ALWAYS TRIGGER
iam_users_profile_backfill_insert_guard`)
	}
	if err != nil {
		t.Fatalf("seed corrupt future profile fixture: %v", err)
	}
	if err = tx.Commit(ctx); err != nil {
		t.Fatalf("commit corrupt future profile fixture: %v", err)
	}
}
