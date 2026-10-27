package application

import (
	"slices"
	"strings"
	"testing"

	fieldregistry "github.com/Onurlulardan/harpy/core/internal/platform/fieldcatalog/fieldregistry"
	presentation "github.com/Onurlulardan/harpy/core/internal/platform/fieldcatalog/presentation"
	layoutapi "github.com/Onurlulardan/harpy/core/internal/modules/system_resource_layout/api"
	"github.com/google/uuid"
)

func TestCuratedOrganizationPresentationOmitsRawAndDuplicateFields(t *testing.T) {
	custom := "842de2f7-a4a9-7605-b7de-3063f65aa092"
	catalog := fixtureCatalogForIDs(t,
		uuid.MustParse(custom),
		uuid.MustParse("c1000000-0000-4000-8000-000000000001"),
		uuid.MustParse(organizationName),
		uuid.MustParse(organizationSlug),
		uuid.MustParse(organizationStatus),
		uuid.MustParse(organizationParentID),
		uuid.MustParse("c1000000-0000-4000-8000-000000000006"),
		uuid.MustParse("c1000000-0000-4000-8000-000000000007"),
		uuid.MustParse("c1000000-0000-4000-8000-000000000008"),
		uuid.MustParse("c1000000-0000-4000-8000-000000000009"),
		uuid.MustParse(organizationCreated),
		uuid.MustParse(organizationUpdated),
	)
	document := mustDefaultPresentation(t, layoutapi.ResourceOrganization, catalog)
	assertPresentationFieldIDs(t, document, "record_create",
		[]string{organizationName, organizationSlug, organizationParentID})
	assertPresentationFieldIDs(t, document, "record_list",
		[]string{
			organizationName, organizationSlug, organizationStatus,
			organizationParentID, organizationCreated,
		})
	for _, surface := range document.Surfaces {
		if len(surface.Translations) != 2 ||
			len(surface.Root.Children[0].Translations) != 2 {
			t.Fatalf("%s lacks en/tr layout metadata", surface.Kind)
		}
		for _, field := range surface.Root.Children[0].Children {
			if field.FieldID == custom || len(field.Translations) != 2 {
				t.Fatalf("%s exposes duplicate/untranslated field %s",
					surface.Kind, field.FieldID)
			}
		}
	}
}

func TestCuratedUserListKeepsDisplayNameAndRelationFields(t *testing.T) {
	custom := uuid.MustParse("301c000e-e7b8-7a4d-a910-8e0a113dcf53")
	catalog := fixtureCatalogForIDs(t, custom,
		uuid.MustParse(userEmail), uuid.MustParse(userStatus),
		uuid.MustParse(userIdentityKind), uuid.MustParse(userRoleIDs),
		uuid.MustParse(userOrganizationIDs), uuid.MustParse(userCreatedAt),
	)
	document := mustDefaultPresentation(t, layoutapi.ResourceUser, catalog)
	assertPresentationFieldIDs(t, document, "record_list", []string{
		custom.String(), userEmail, userStatus, userIdentityKind,
		userOrganizationIDs, userRoleIDs, userCreatedAt,
	})
	assertPresentationFieldIDs(t, document, "record_update", []string{
		custom.String(), userOrganizationIDs, userRoleIDs,
	})
}

func TestCuratedMutationSurfacesPreserveSupportedLegacyFields(t *testing.T) {
	t.Run("user create omits non-mutation portal identifier", func(t *testing.T) {
		custom := "301c000e-e7b8-7a4d-a910-8e0a113dcf53"
		catalog := fixtureCatalogWithModes(t, map[string]presentation.FieldMode{
			custom:               presentation.FieldEditable,
			userEmail:            presentation.FieldEditable,
			userPortalIdentifier: presentation.FieldReadOnly,
			userOrganizationIDs:  presentation.FieldEditable,
			userRoleIDs:          presentation.FieldEditable,
			userStatus:           presentation.FieldReadOnly,
		})
		document := mustDefaultPresentation(t, layoutapi.ResourceUser, catalog)
		assertPresentationFieldIDs(t, document, "record_create", []string{
			custom, userEmail, userOrganizationIDs, userRoleIDs,
		})
	})

	t.Run("role create retains default selection", func(t *testing.T) {
		catalog := fixtureCatalogWithModes(t, map[string]presentation.FieldMode{
			roleName:           presentation.FieldEditable,
			roleDescription:    presentation.FieldEditable,
			roleScopeKind:      presentation.FieldEditable,
			roleOrganizationID: presentation.FieldEditable,
			roleDefault:        presentation.FieldEditable,
		})
		document := mustDefaultPresentation(t, layoutapi.ResourceRole, catalog)
		assertPresentationFieldIDs(t, document, "record_create", []string{
			roleName, roleDescription, roleScopeKind, roleOrganizationID, roleDefault,
		})
	})

	t.Run("organization owner follows IAM mutability", func(t *testing.T) {
		catalog := fixtureCatalogWithModes(t, map[string]presentation.FieldMode{
			organizationName:     presentation.FieldEditable,
			organizationSlug:     presentation.FieldEditable,
			organizationParentID: presentation.FieldEditable,
			organizationOwnerIDs: presentation.FieldEditable,
			organizationStatus:   presentation.FieldReadOnly,
		})
		document := mustDefaultPresentation(t, layoutapi.ResourceOrganization, catalog)
		want := []string{
			organizationName, organizationSlug, organizationParentID,
			organizationOwnerIDs,
		}
		assertPresentationFieldIDs(t, document, "record_create", want)
		assertPresentationFieldIDs(t, document, "record_update", want)
	})
}

func fixtureCatalogWithModes(
	t *testing.T,
	fieldsByID map[string]presentation.FieldMode,
) presentation.FieldCatalog {
	t.Helper()
	fields := make([]presentation.Field, 0, len(fieldsByID))
	for id, mode := range fieldsByID {
		field, err := presentation.NewField(
			id, fieldregistry.TypeText, fieldregistry.ValueTextNFCV1, mode,
		)
		if err != nil {
			t.Fatal(err)
		}
		fields = append(fields, field)
	}
	slices.SortFunc(fields, func(left, right presentation.Field) int {
		leftID, _ := left.ID()
		rightID, _ := right.ID()
		return strings.Compare(leftID, rightID)
	})
	catalog, err := presentation.NewFieldCatalog(fields)
	if err != nil {
		t.Fatal(err)
	}
	return catalog
}

func mustDefaultPresentation(
	t *testing.T,
	resource layoutapi.ResourceKind,
	catalog presentation.FieldCatalog,
) presentation.Document {
	t.Helper()
	document, err := defaultPresentation(resource, catalog)
	if err != nil {
		t.Fatal(err)
	}
	if len(document.Locales) != 2 ||
		document.Locales[0] != "en" || document.Locales[1] != "tr" {
		t.Fatalf("locales=%v", document.Locales)
	}
	return document
}

func assertPresentationFieldIDs(
	t *testing.T,
	document presentation.Document,
	kind string,
	expected []string,
) {
	t.Helper()
	for _, surface := range document.Surfaces {
		if surface.Kind != kind {
			continue
		}
		fields := surface.Root.Children[0].Children
		if len(fields) != len(expected) {
			t.Fatalf("%s fields=%d expected=%d", kind, len(fields), len(expected))
		}
		for index, field := range fields {
			if field.FieldID != expected[index] {
				t.Fatalf("%s field[%d]=%s expected=%s",
					kind, index, field.FieldID, expected[index])
			}
		}
		return
	}
	t.Fatalf("surface %s absent", kind)
}
