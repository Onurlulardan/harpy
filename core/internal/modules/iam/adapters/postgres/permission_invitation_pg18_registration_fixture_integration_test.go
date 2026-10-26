//go:build integration

package postgres

import (
	"bytes"
	"context"
	"crypto/sha256"
	"database/sql"
	"strings"
	"testing"
)

const invitationDeliveryKMSKey = "gcp-kms://projects/harpy-prod/locations/" +
	"europe-west1/keyRings/iam/cryptoKeys/delivery/cryptoKeyVersions/7"

type iamInvitationRegistrationFixture struct {
	invitationID, tokenID, artifactID, createOperationID string
	acceptOperationID, userID, identifierID              string
	accessModeID, assignmentID, email                    string
	selector, secretHMAC, binding                        []byte
}

func newIAMInvitationRegistrationFixture(
	tenant iamFoundationTenant, marker string,
) iamInvitationRegistrationFixture {
	return iamInvitationRegistrationFixture{
		invitationID:      iamPermissionInvitationID("d1", marker),
		tokenID:           iamPermissionInvitationID("d2", marker),
		artifactID:        iamPermissionInvitationID("d3", marker),
		createOperationID: iamPermissionInvitationID("d4", marker),
		acceptOperationID: iamPermissionInvitationID("d5", marker),
		userID:            iamPermissionInvitationID("d6", marker),
		identifierID:      iamPermissionInvitationID("d7", marker),
		accessModeID:      iamPermissionInvitationID("d8", marker),
		assignmentID:      iamPermissionInvitationID("d9", marker),
		email:             "registered-" + marker + "@example.com",
		selector:          bytes.Repeat([]byte{0x31}, 16),
		secretHMAC:        bytes.Repeat([]byte{0x32}, 32),
		binding:           bytes.Repeat([]byte{0x33}, 32),
	}
}

func createIAMInvitationRegistrationV2(
	t *testing.T, ctx context.Context, database *sql.DB,
	tenant iamFoundationTenant, fixture iamInvitationRegistrationFixture,
) {
	t.Helper()
	artifactRef := "iam-delivery://tenant/" + tenant.id +
		"/artifact/" + fixture.artifactID
	result := mustRunIAMPGFirstText(
		t, ctx, database, tenant.schema, iamInvitationCreateV2SQL,
		fixture.createOperationID, bytes.Repeat([]byte{0x41}, 32),
		fixture.invitationID, fixture.email, fixture.email,
		bytes.Repeat([]byte{0x42}, 32), iamInvitationRole, tenant.id,
		iamInvitationManager, fixture.tokenID, fixture.selector,
		strings.Repeat("h", 16), fixture.secretHMAC, fixture.binding,
		fixture.artifactID, artifactRef, invitationDeliveryKMSKey,
		bytes.Repeat([]byte{0x43}, 12), bytes.Repeat([]byte{0x44}, 66),
		bytes.Repeat([]byte{0x45}, 16), bytes.Repeat([]byte{0x46}, 32),
		bytes.Repeat([]byte{0x47}, 32),
	)
	if result != fixture.invitationID+"|pending|1|false" {
		t.Fatalf("invitation V2 create=%q", result)
	}
}

func invitationRegistrationArguments(
	fixture iamInvitationRegistrationFixture,
) ([]any, []any) {
	profile := []byte("{}")
	profileSHA := sha256.Sum256(profile)
	registration := []any{
		fixture.acceptOperationID, bytes.Repeat([]byte{0x51}, 32),
		fixture.invitationID, fixture.tokenID, fixture.selector,
		strings.Repeat("h", 16), fixture.secretHMAC, fixture.binding,
		fixture.userID, fixture.identifierID, fixture.accessModeID,
		fixture.assignmentID, nil, iamUserProfileSchemaID, profile,
		profileSHA[:], bytes.Repeat([]byte{0x52}, 16), bytes.Repeat([]byte{0x53}, 32),
	}
	directorySHA := sha256.Sum256([]byte(fixture.email))
	directory := []any{
		fixture.userID, fixture.identifierID, strings.Repeat("d", 16),
		bytes.Repeat([]byte{0x54}, 32), directorySHA[:],
	}
	return registration, directory
}

const iamInvitationCreateV2SQL = `
SELECT pg_catalog.concat_ws('|',invitation_id::text,invitation_state,
    invitation_revision::text,replayed::text)
FROM iam_invitation_create_v2(
    (pg_catalog.current_schemas(false))[2]::name,
    $1,$2,$3,$4,$5,$6,$7,'tenant',$8,NULL,$9,1,$10,$11,$12,$13,$14,
    $15,$16,$17,$18,$19,$20,$21,$22
)`

const iamInvitationRegisterSQL = `
SELECT pg_catalog.concat_ws('|',result_code,result_revision::text,
    authorization_clock::text,replayed::text)
FROM iam_invitation_register_accept_v1(
    (pg_catalog.current_schemas(false))[2]::name,
    $1,$2,$3,1,$4,1,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,1
)`

const iamInvitationDirectorySQL = `
SELECT platform.iam_identity_directory_upsert_v2(
    $1::uuid,1,1,$2::uuid,$3::uuid,$4::text,$5::bytea,1,$6::bytea
)`
