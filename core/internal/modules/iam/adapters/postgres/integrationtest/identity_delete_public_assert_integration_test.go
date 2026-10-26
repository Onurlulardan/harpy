//go:build integration

package integrationtest

import (
	"context"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/profile"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/restquery/catalog"
)

func TestPublicationPublicDeleteCatalogFixture(t *testing.T) {
	custom, err := profile.NewQueryCatalog(publicationCurrentSchema(t))
	if err != nil {
		t.Fatal(err)
	}
	if _, err = catalog.Build(custom); err != nil {
		t.Fatalf("public delete user catalog fixture: %v", err)
	}
}

func assertPublicationPublicUserDelete(
	t *testing.T, ctx context.Context, db publicationDatabase, runtime publicationRuntime,
) {
	t.Helper()
	port := publicationIdentityPort(t, runtime)
	first := callPublicationPublicDelete(t, ctx, db, runtime, port)
	firstReceipt, err := first.Receipt()
	if err != nil {
		t.Fatal(err)
	}
	firstRevision, firstReplayed, err := firstReceipt.Result()
	if err != nil || firstRevision != 4 || firstReplayed {
		t.Fatalf("public delete receipt = %d/%t/%v",
			firstRevision, firstReplayed, err)
	}
	assertPublicationPublicDeleteState(t, ctx, db, firstRevision)
	replay := callPublicationPublicDelete(t, ctx, db, runtime, port)
	replayReceipt, err := replay.Receipt()
	if err != nil {
		t.Fatal(err)
	}
	replayRevision, replayed, err := replayReceipt.Result()
	if err != nil || replayRevision != firstRevision || !replayed {
		t.Fatalf("public delete replay = %d/%t/%v",
			replayRevision, replayed, err)
	}
	firstOperation, _ := firstReceipt.OperationID()
	replayOperation, _ := replayReceipt.OperationID()
	firstSHA, _ := firstReceipt.RequestSHA256()
	replaySHA, _ := replayReceipt.RequestSHA256()
	if firstOperation != replayOperation || firstSHA != replaySHA {
		t.Fatal("public delete replay receipt changed")
	}
	assertPublicationPublicDeleteState(t, ctx, db, firstRevision)
}

func assertPublicationPublicDeleteState(
	t *testing.T, ctx context.Context, db publicationDatabase, revision int64,
) {
	t.Helper()
	var userState, identifierState, accessState string
	var membershipState, assignmentState, directoryState string
	var userRevision, directoryRevision int64
	err := db.admin.QueryRow(ctx, `
SELECT usr.status,usr.revision,identifier.state,access.state,
 membership.state,assignment.state,directory.state,directory.source_revision
FROM t_profile_publication.iam_users AS usr
JOIN t_profile_publication.iam_user_identifiers AS identifier
 ON identifier.id=$2
JOIN t_profile_publication.iam_user_access_modes AS access ON access.id=$3
JOIN t_profile_publication.iam_organization_memberships AS membership
 ON membership.id=$4
JOIN t_profile_publication.iam_user_role_assignments AS assignment
 ON assignment.id=$5
JOIN platform.iam_login_directory AS directory
 ON directory.tenant_id=$6 AND directory.user_id=usr.id
WHERE usr.id=$1`, publicDeleteUserID, publicDeleteIdentifierID,
		publicDeleteAccessID, publicDeleteMembershipID, publicDeleteAssignmentID,
		db.tenant.id).Scan(
		&userState, &userRevision, &identifierState, &accessState,
		&membershipState, &assignmentState, &directoryState, &directoryRevision,
	)
	if err != nil || userState != "tombstoned" || userRevision != revision ||
		identifierState != "tombstoned" || accessState != "retired" ||
		membershipState != "tombstoned" || assignmentState != "revoked" ||
		directoryState != "tombstoned" || directoryRevision != revision {
		t.Fatalf("public delete state = %s/%d/%s/%s/%s/%s/%s/%d: %v",
			userState, userRevision, identifierState, accessState, membershipState,
			assignmentState, directoryState, directoryRevision, err)
	}
}
