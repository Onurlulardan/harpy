//go:build integration

package integrationtest

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"testing"

	profilepostgres "github.com/Onurlulardan/harpy/core/internal/modules/iam/adapters/postgres/identity/profile"
	userpostgres "github.com/Onurlulardan/harpy/core/internal/modules/iam/adapters/postgres/identity/user"
	userapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/identity/user"
	identitydomain "github.com/Onurlulardan/harpy/core/internal/modules/iam/domain/identity"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

const publicationWriterOperationID = "54000000-0000-4000-8000-000000000001"

var publicationNicknamePatch = []byte(
	`{"43000000-0000-4000-8000-000000000002":"nick"}`,
)

func assertPublicationHistoricalNormalWrite(
	t *testing.T, ctx context.Context, db publicationDatabase,
	runtime publicationRuntime, planID uuid.UUID,
) {
	t.Helper()
	seedPublicationSelfIdentifier(t, ctx, db)
	assertPublicationProfileSourceRead(t, ctx, db, runtime)
	claims := publicationClaims(t, db.tenant)
	beforeView, err := runtime.application.GetSelf(ctx, claims)
	if err != nil {
		t.Fatalf("read historical profile through current source: %v", err)
	}
	historical := publicationViewString(t, beforeView.Document, "display_name")
	if len([]byte(historical)) <= 8 ||
		publicationViewString(t, beforeView.Document, "nickname") != "" {
		t.Fatalf("historical drift fixture = %q/%s", historical, beforeView.Document)
	}
	var sourceRevision int64
	var sourceSHA []byte
	if err = db.admin.QueryRow(ctx, `SELECT revision,profile_sha256
FROM t_profile_publication.iam_users WHERE id=$1`, publicationActorID).
		Scan(&sourceRevision, &sourceSHA); err != nil {
		t.Fatal(err)
	}
	result := runPublicationNormalWrite(
		t, ctx, db, runtime, sourceRevision, publicationNicknamePatch,
	)
	if result.ID != identitydomain.IdentityID(uuid.MustParse(publicationActorID)) ||
		result.Revision != sourceRevision+1 || result.Replayed {
		t.Fatalf("normal writer result = %#v", result)
	}
	assertPublicationPreconversion(
		t, ctx, db, planID, sourceRevision, sourceSHA, historical,
	)
	afterView, err := runtime.application.GetSelf(ctx, claims)
	if err != nil || afterView.Revision != result.Revision ||
		publicationViewString(t, afterView.Document, "display_name") != historical ||
		publicationViewString(t, afterView.Document, "nickname") != "nick" {
		t.Fatalf("read migrated profile = %d/%s/%v",
			afterView.Revision, afterView.Document, err)
	}
}

func runPublicationNormalWrite(
	t *testing.T, ctx context.Context, db publicationDatabase,
	runtime publicationRuntime, expectedRevision int64, patch []byte,
) identitydomain.MutationResult {
	t.Helper()
	profiles, err := profilepostgres.NewStore()
	if err != nil {
		t.Fatal(err)
	}
	users, err := userpostgres.NewStore(profiles)
	if err != nil {
		t.Fatal(err)
	}
	commands, err := userapp.NewCommands(users, profiles)
	if err != nil {
		t.Fatal(err)
	}
	tenantID := uuid.MustParse(db.tenant.id)
	generations, err := runtime.pool.ResolveSnapshotDomainGenerations(
		ctx, tenantID, []string{"iam", "system_resource_layout"},
	)
	if err != nil {
		t.Fatal(err)
	}
	request, err := platformtx.AdmitCommandScope(
		uuid.MustParse(publicationActorID), tenantID, "iam.identity.user.update",
		"iam_system_resources", generations["iam"], generations,
	)
	if err != nil {
		t.Fatal(err)
	}
	evidence := identitydomain.MutationEvidence{
		OperationID:          identitydomain.IdentityID(uuid.MustParse(publicationWriterOperationID)),
		RequestSHA256:        sha256.Sum256(patch),
		ActorID:              identitydomain.IdentityID(uuid.MustParse(publicationActorID)),
		PermissionID:         identitydomain.IdentityID(uuid.MustParse(publicationManagePerm)),
		ExpectedTokenVersion: 1, ExpectedAuthorizationVersion: 1,
	}
	var result identitydomain.MutationResult
	err = runtime.coordinator.InTenant(
		ctx, request, func(scoped context.Context, scope platformtx.TenantScope) error {
			if aliasErr := assertPublicationHistoricalAlias(
				scoped, scope, profiles,
			); aliasErr != nil {
				return aliasErr
			}
			result, err = commands.UpdateUser(scoped, scope, userapp.UpdateUserCommand{
				ID:               identitydomain.IdentityID(uuid.MustParse(publicationActorID)),
				ExpectedRevision: expectedRevision, ProfilePatch: patch, Evidence: evidence,
			})
			return err
		},
	)
	if err != nil {
		t.Fatalf("normal profile writer: %v", err)
	}
	return result
}

func assertPublicationPreconversion(
	t *testing.T, ctx context.Context, db publicationDatabase, planID uuid.UUID,
	sourceRevision int64, sourceSHA []byte, historical string,
) {
	t.Helper()
	var schema uuid.UUID
	var revision, receiptSourceRevision, firstTargetRevision int64
	var oldValue, nickname, replacement string
	var receiptSourceSHA, plannedSHA, firstSHA, rowSHA []byte
	err := db.admin.QueryRow(ctx, `SELECT usr.profile_schema_version_id,usr.revision,
usr.profile_values->>$3::text,usr.profile_values->>$4::text,
usr.profile_values->>$5::text,usr.profile_sha256,receipt.source_revision,
receipt.first_target_revision,receipt.source_sha256,
receipt.planned_target_sha256,receipt.first_target_sha256
FROM t_profile_publication.iam_users AS usr
JOIN t_profile_publication.iam_profile_backfill_preconversions AS receipt
 ON receipt.plan_id=$1 AND receipt.resource_id=usr.id WHERE usr.id=$2`,
		planID, publicationActorID, publicationSourceField,
		publicationOptionalID, publicationReplacement).Scan(
		&schema, &revision, &oldValue, &nickname, &replacement, &rowSHA,
		&receiptSourceRevision, &firstTargetRevision, &receiptSourceSHA,
		&plannedSHA, &firstSHA,
	)
	if err != nil || schema != uuid.MustParse(publicationTargetV3ID) ||
		revision != sourceRevision+1 || oldValue != historical || nickname != "nick" ||
		replacement == "" || receiptSourceRevision != sourceRevision ||
		firstTargetRevision != sourceRevision+1 || !bytes.Equal(receiptSourceSHA, sourceSHA) ||
		bytes.Equal(plannedSHA, firstSHA) || !bytes.Equal(firstSHA, rowSHA) {
		t.Fatalf("preconversion = %s/%d/%q/%q/%q/%d/%d/%v",
			schema, revision, oldValue, nickname, replacement,
			receiptSourceRevision, firstTargetRevision, err)
	}
}

func publicationViewString(t *testing.T, document []byte, key string) string {
	t.Helper()
	var values map[string]json.RawMessage
	if json.Unmarshal(document, &values) != nil || len(values[key]) == 0 ||
		bytes.Equal(values[key], []byte("null")) {
		return ""
	}
	var value string
	if json.Unmarshal(values[key], &value) != nil {
		t.Fatalf("profile view field %q = %s", key, values[key])
	}
	return value
}

func seedPublicationSelfIdentifier(
	t *testing.T, ctx context.Context, db publicationDatabase,
) {
	t.Helper()
	const statement = `INSERT INTO iam_user_identifiers(
id,user_id,kind,display_value,normalized_value,normalized_sha256,state,revision,
verified_at,activated_at,tombstoned_at,cooldown_until,retired_at,created_at,updated_at)
VALUES('55000000-0000-4000-8000-000000000001',$1,'email_v1',$2,$2,
pg_catalog.sha256(pg_catalog.convert_to($2,'UTF8')),'active',1,
pg_catalog.now(),pg_catalog.now(),NULL,NULL,NULL,pg_catalog.now(),pg_catalog.now())`
	execPublicationTenant(t, ctx, db, statement,
		publicationActorID, "publisher@example.com")
}
