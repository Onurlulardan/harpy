package systemresource

import (
	"crypto/sha256"
	"testing"
	"time"

	publicresource "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/systemresource"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/profile"
	identitydomain "github.com/Onurlulardan/harpy/core/internal/modules/iam/domain/identity"
	"github.com/google/uuid"
)

func TestUserRecordProjectsFrozenWireContract(t *testing.T) {
	catalog := wireProjectionCatalog(t, profile.ResourceUser)
	now := time.Date(2026, 8, 2, 10, 11, 12, 123456000, time.UTC)
	organizationID := wireIdentity("71000000-0000-4000-8000-000000000001")
	user := identitydomain.User{
		ID:     wireIdentity("71000000-0000-4000-8000-000000000002"),
		Status: identitydomain.UserSuspended, TokenVersion: 1,
		AuthorizationVersion: 2, Revision: 3,
		IdentifierID:    wireIdentity("71000000-0000-4000-8000-000000000003"),
		EmailDisplay:    "user@example.com",
		EmailNormalized: "user@example.com",
		RoleIDs:         []identitydomain.IdentityID{},
		OrganizationIDs: []identitydomain.IdentityID{organizationID},
		EmailVerified:   true,
		AccessModes:     []string{"operator_console", "portal"},
		Profile:         wireProfile(t, identitydomain.ProfileUser),
		CreatedAt:       now, UpdatedAt: now.Add(time.Second),
	}
	record, err := UserRecord(user, catalog)
	if err != nil {
		t.Fatal(err)
	}
	assertWireMember(t, record, catalog, publicresource.UserStatusFieldID,
		"status", `"INACTIVE"`)
	assertWireMember(t, record, catalog, publicresource.UserEmailVerifiedFieldID,
		"email_verified", `true`)
	assertWireMember(t, record, catalog, publicresource.UserAccountTypeFieldID,
		"account_type", `"operator_portal"`)
	assertWireMember(t, record, catalog, publicresource.UserOrganizationIDFieldID,
		"organization_id", `"`+uuid.UUID(organizationID).String()+`"`)

	user.OrganizationIDs = append(user.OrganizationIDs,
		wireIdentity("71000000-0000-4000-8000-000000000004"))
	record, err = UserRecord(user, catalog)
	if err != nil {
		t.Fatal(err)
	}
	assertWireMemberAbsent(t, record, publicresource.UserOrganizationIDFieldID)
}

func TestRoleAndOrganizationRecordsProjectFrozenStatesAndOwners(t *testing.T) {
	now := time.Date(2026, 8, 2, 10, 11, 12, 123456000, time.UTC)
	roleCatalog := wireProjectionCatalog(t, profile.ResourceRole)
	role := identitydomain.Role{
		ID:        wireIdentity("72000000-0000-4000-8000-000000000001"),
		ScopeKind: identitydomain.RoleScopeTenant,
		ScopeID:   wireIdentity("72000000-0000-4000-8000-000000000002"),
		Name:      "Auditor", Status: identitydomain.RoleRetired, Revision: 2,
		Profile:   wireProfile(t, identitydomain.ProfileRole),
		CreatedAt: now, UpdatedAt: now.Add(time.Second), RetiredAt: now.Add(time.Second),
	}
	roleRecord, err := RoleRecord(role, roleCatalog)
	if err != nil {
		t.Fatal(err)
	}
	assertWireMember(t, roleRecord, roleCatalog, publicresource.RoleStatusFieldID,
		"status", `"RETIRED"`)

	organizationCatalog := wireProjectionCatalog(t, profile.ResourceOrganization)
	ownerID := wireIdentity("73000000-0000-4000-8000-000000000001")
	organization := identitydomain.Organization{
		ID:   wireIdentity("73000000-0000-4000-8000-000000000002"),
		Name: "Archive", Slug: "archive", Status: identitydomain.OrganizationTombstoned,
		OwnerIDs: []identitydomain.IdentityID{ownerID}, Revision: 4,
		Profile:   wireProfile(t, identitydomain.ProfileOrganization),
		CreatedAt: now, UpdatedAt: now.Add(time.Second), TombstonedAt: now.Add(time.Second),
	}
	organizationRecord, err := OrganizationRecord(organization, organizationCatalog)
	if err != nil {
		t.Fatal(err)
	}
	assertWireMember(t, organizationRecord, organizationCatalog,
		publicresource.OrganizationStatusFieldID, "status", `"DELETED"`)
	assertWireMember(t, organizationRecord, organizationCatalog,
		publicresource.OrganizationOwnerIDsFieldID, "owner_ids",
		`["`+uuid.UUID(ownerID).String()+`"]`)
	assertWireMember(t, organizationRecord, organizationCatalog,
		publicresource.OrganizationOwnerIDFieldID, "owner_id",
		`"`+uuid.UUID(ownerID).String()+`"`)
}

func wireProjectionCatalog(
	t *testing.T, kind profile.ResourceKind,
) publicresource.FieldCatalog {
	t.Helper()
	visibility, err := NewCatalogVisibility(true, true, true)
	if err != nil {
		t.Fatal(err)
	}
	catalog, err := BuildFieldCatalog(publishedFixture(t, kind, nil), visibility)
	if err != nil {
		t.Fatal(err)
	}
	return catalog
}

func wireProfile(
	t *testing.T, resource identitydomain.ProfileResource,
) identitydomain.ProfileDocument {
	t.Helper()
	values := []byte(`{}`)
	document, err := identitydomain.NewProfileDocument(
		resource, wireIdentity("74000000-0000-4000-8000-000000000001"),
		values, sha256.Sum256(values),
	)
	if err != nil {
		t.Fatal(err)
	}
	return document
}

func wireIdentity(value string) identitydomain.IdentityID {
	return identitydomain.IdentityID(uuid.MustParse(value))
}

func assertWireMember(
	t *testing.T,
	record publicresource.Record,
	catalog publicresource.FieldCatalog,
	fieldID string,
	wantName string,
	wantValue string,
) {
	t.Helper()
	id, err := publicresource.ParseID(fieldID)
	if err != nil {
		t.Fatal(err)
	}
	descriptor, found, err := catalog.Descriptor(id)
	if err != nil || !found {
		t.Fatalf("descriptor %s missing: %v", fieldID, err)
	}
	name, err := descriptor.PublicName()
	if err != nil || name != wantName {
		t.Fatalf("descriptor name = %q/%v", name, err)
	}
	values, err := record.Values()
	if err != nil {
		t.Fatal(err)
	}
	members, err := values.Values()
	if err != nil || string(members[id]) != wantValue {
		t.Fatalf("%s = %s/%v, want %s", wantName, members[id], err, wantValue)
	}
}

func assertWireMemberAbsent(
	t *testing.T, record publicresource.Record, fieldID string,
) {
	t.Helper()
	id, _ := publicresource.ParseID(fieldID)
	values, _ := record.Values()
	members, _ := values.Values()
	if _, found := members[id]; found {
		t.Fatalf("field %s should be absent", fieldID)
	}
}
