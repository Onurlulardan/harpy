package systemresource

import (
	"bytes"
	"crypto/sha256"
	"testing"
	"time"

	canonicaljsonv1 "github.com/Onurlulardan/harpy/contracts/canonicaljson/v1"
	publicresource "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/systemresource"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/profile"
	identitydomain "github.com/Onurlulardan/harpy/core/internal/modules/iam/domain/identity"
	"github.com/google/uuid"
)

func TestUserRecordUsesStableFieldsAndVisibleProfileOnly(t *testing.T) {
	customID := uuid.MustParse("e1000000-0000-4000-8000-000000000002")
	custom, err := profile.NewField(profile.FieldInput{
		ID: customID, Resource: profile.ResourceUser, Key: "private_note",
		Type: profile.LogicalString, Cardinality: profile.CardinalitySingle,
		Classification: profile.ClassificationRestricted,
		Normalization:  profile.NormalizationNoneV1, State: profile.FieldActive,
	})
	if err != nil {
		t.Fatal(err)
	}
	published := publishedFixture(t, profile.ResourceUser, []profile.Field{custom})
	visibility, _ := NewCatalogVisibility(true, true, false)
	catalog, err := BuildFieldCatalog(published, visibility)
	if err != nil {
		t.Fatal(err)
	}
	profileJSON, _ := canonicaljsonv1.CanonicalizeObject(
		[]byte(`{"e1000000-0000-4000-8000-000000000002":"hidden"}`), 1<<20,
	)
	profileDocument, err := identitydomain.NewProfileDocument(
		identitydomain.ProfileUser,
		identitydomain.IdentityID(uuid.MustParse("f1000000-0000-4000-8000-000000000001")),
		profileJSON, sha256.Sum256(profileJSON),
	)
	if err != nil {
		t.Fatal(err)
	}
	userID := uuid.MustParse("018f0f4e-7b61-7c25-8000-000000000041")
	createdAt := time.Date(2026, 7, 17, 12, 13, 14, 987654321, time.UTC)
	user := identitydomain.User{
		ID: identitydomain.IdentityID(userID), Status: identitydomain.UserActive,
		TokenVersion: 1, AuthorizationVersion: 2, Revision: 3,
		IdentifierID: identitydomain.IdentityID(
			uuid.MustParse("018f0f4e-7b61-7c25-8000-000000000042"),
		),
		EmailDisplay: "user@example.com", EmailNormalized: "user@example.com",
		RoleIDs: []identitydomain.IdentityID{}, OrganizationIDs: []identitydomain.IdentityID{},
		AccessModes: []string{},
		Profile:     profileDocument, CreatedAt: createdAt,
		UpdatedAt: createdAt.Add(2 * time.Second),
	}
	record, err := UserRecord(user, catalog)
	if err != nil {
		t.Fatal(err)
	}
	values, _ := record.Values()
	members, _ := values.Values()
	if _, leaked := members[publicresource.ID(customID)]; leaked {
		t.Fatal("restricted profile value leaked")
	}
	emailID := publicresource.ID(uuid.MustParse(publicresource.UserEmailFieldID))
	if string(members[emailID]) != `"user@example.com"` {
		t.Fatalf("email=%s", members[emailID])
	}
	createdID := publicresource.ID(uuid.MustParse(publicresource.UserCreatedAtFieldID))
	if string(members[createdID]) != `"2026-07-17T12:13:14.987654Z"` {
		t.Fatalf("created_at=%s", members[createdID])
	}
	canonical, _ := values.CanonicalJSON()
	if bytes.Contains(canonical, []byte("password")) {
		t.Fatal("password value entered canonical record")
	}
}
