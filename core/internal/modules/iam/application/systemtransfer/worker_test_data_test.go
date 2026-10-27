package systemtransfer

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"testing"

	tabulartransfer "github.com/Onurlulardan/harpy/core/internal/capabilities/tabulartransfer/api"
	publicresource "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/systemresource"
	"github.com/google/uuid"
)

var (
	fixtureTransferID = uuid.MustParse("019be000-0000-7000-8000-000000000001")
	fixtureTenantID   = uuid.MustParse("019be000-0000-7000-8000-000000000002")
	fixtureActorID    = uuid.MustParse("019be000-0000-7000-8000-000000000003")
	fixtureRecordID   = uuid.MustParse("019be000-0000-7000-8000-000000000004")
	fixtureRoleID     = uuid.MustParse("019be000-0000-7000-8000-000000000005")
	fixtureInviteID   = uuid.MustParse("019be000-0000-7000-8000-000000000006")
	fixtureJobID      = uuid.MustParse("019be000-0000-7000-8000-000000000007")
	fixtureAttemptID  = uuid.MustParse("019be000-0000-7000-8000-000000000008")
	fixtureLeaseID    = uuid.MustParse("019be000-0000-7000-8000-000000000009")
	fixtureCatalogID  = uuid.MustParse("019be000-0000-7000-8000-000000000010")
)

func fixtureAttempt(t *testing.T, resource publicresource.Resource) Attempt {
	t.Helper()
	policy, err := newMappingPolicy(resource, nil)
	if err != nil {
		t.Fatal(err)
	}
	document := fixtureMappingDocument(resource)
	program, err := policy.Compile([]byte(document))
	if err != nil {
		t.Fatal(err)
	}
	manifest, err := ManifestFor(resource)
	if err != nil {
		t.Fatal(err)
	}
	return Attempt{ID: fixtureTransferID, JobID: fixtureJobID,
		AttemptID: fixtureAttemptID, LeaseOwnerID: fixtureLeaseID, LeaseGeneration: 1,
		TenantID: fixtureTenantID,
		ActorID:  fixtureActorID, Resource: resource, AuthorizationVersion: 7,
		CatalogVersionID:       fixtureCatalogID,
		CatalogSHA256:          sha256.Sum256([]byte("catalog-v1")),
		IdentityManifestSHA256: manifest.Digest,
		DuplicateAction:        DuplicateUpdate, Mapping: program}
}

func fixtureMappingDocument(resource publicresource.Resource) string {
	fields := ""
	switch resource {
	case publicresource.Users:
		fields = fixtureDirectField(publicresource.UserEmailFieldID, "email") + "," +
			fixtureDirectField(publicresource.UserRoleIDsFieldID, "role_ids")
	case publicresource.Roles:
		fields = fixtureDirectField(publicresource.RoleScopeKindFieldID, "scope_kind") + "," +
			fixtureDirectField(publicresource.RoleNameFieldID, "name")
	case publicresource.Organizations:
		fields = fixtureDirectField(publicresource.OrganizationNameFieldID, "name") + "," +
			fixtureDirectField(publicresource.OrganizationSlugFieldID, "slug")
	}
	return `{"contract":"tabular-mapping.v2","fields":[` + fields + `]}`
}

func fixtureDirectField(target, source string) string {
	return fmt.Sprintf(`{"target":"%s","required":true,`+
		`"expression":{"kind":"direct","source":"%s"}}`, target, source)
}

func fixtureSourceRow(t *testing.T, resource publicresource.Resource) SourceRow {
	t.Helper()
	document := ""
	switch resource {
	case publicresource.Users:
		document = fmt.Sprintf(`{"email":"person@example.com","role_ids":["%s"]}`,
			fixtureRoleID)
	case publicresource.Roles:
		document = `{"scope_kind":"tenant","name":"operators"}`
	case publicresource.Organizations:
		document = `{"name":"Operations","slug":"operations"}`
	}
	reader, err := tabulartransfer.NewReader(tabulartransfer.FormatJSONL,
		bytes.NewBufferString(document+"\n"), tabulartransfer.DefaultLimits())
	if err != nil {
		t.Fatal(err)
	}
	rows, done, err := ReadSourceBatch(reader, tabulartransfer.FormatJSONL, 1)
	if err != nil || len(rows) != 1 || done {
		t.Fatalf("rows=%d done=%v err=%v", len(rows), done, err)
	}
	return rows[0]
}

func fixtureMatch(found bool, revision int64) IdentityMatch {
	if !found {
		return IdentityMatch{}
	}
	return IdentityMatch{Found: true, ID: publicresource.ID(fixtureRecordID),
		Revision: revision}
}
