package application

import (
	"context"
	"testing"
	"time"

	fieldregistry "github.com/Onurlulardan/harpy/core/internal/platform/fieldcatalog/fieldregistry"
	presentation "github.com/Onurlulardan/harpy/core/internal/platform/fieldcatalog/presentation"
	layoutapi "github.com/Onurlulardan/harpy/core/internal/modules/system_resource_layout/api"
	"github.com/google/uuid"
)

var (
	fixtureRevisionID = uuid.MustParse("11111111-1111-7111-8111-111111111111")
	fixtureSectionID  = uuid.MustParse("22222222-2222-4222-8222-222222222222")
	fixtureFieldOne   = uuid.MustParse("33333333-3333-4333-8333-333333333333")
	fixtureFieldTwo   = uuid.MustParse("44444444-4444-4444-8444-444444444444")
	fixtureFieldThree = uuid.MustParse("55555555-5555-4555-8555-555555555555")
	fixtureNow        = time.Date(2026, 7, 16, 12, 0, 0, 0, time.UTC)
)

type fixtureIDs struct{ id uuid.UUID }

func (source fixtureIDs) NewUUID() (uuid.UUID, error) { return source.id, nil }

type fixtureClock struct{ value time.Time }

func (clock fixtureClock) Now() time.Time { return clock.value }

type fixtureStore struct {
	create  func(context.Context, CreateStoreInput) (layoutapi.Revision, error)
	get     func(context.Context, uuid.UUID) (layoutapi.Revision, error)
	list    func(context.Context, layoutapi.ResourceKind) ([]layoutapi.Revision, error)
	update  func(context.Context, UpdateStoreInput) (layoutapi.Revision, error)
	delete  func(context.Context, uuid.UUID, int64) error
	publish func(context.Context, PublishStoreInput) (PublishedPresentation, error)
	resolve func(context.Context, layoutapi.ResourceKind) (PublishedPresentation, error)
}

func (store *fixtureStore) Create(
	ctx context.Context, input CreateStoreInput,
) (layoutapi.Revision, error) {
	return store.create(ctx, input)
}

func (store *fixtureStore) Get(
	ctx context.Context, id uuid.UUID,
) (layoutapi.Revision, error) {
	return store.get(ctx, id)
}

func (store *fixtureStore) List(
	ctx context.Context, resource layoutapi.ResourceKind,
) ([]layoutapi.Revision, error) {
	return store.list(ctx, resource)
}

func (store *fixtureStore) Update(
	ctx context.Context, input UpdateStoreInput,
) (layoutapi.Revision, error) {
	return store.update(ctx, input)
}

func (store *fixtureStore) Delete(
	ctx context.Context, id uuid.UUID, expected int64,
) error {
	return store.delete(ctx, id, expected)
}

func (store *fixtureStore) Publish(
	ctx context.Context, input PublishStoreInput,
) (PublishedPresentation, error) {
	return store.publish(ctx, input)
}

func (store *fixtureStore) Resolve(
	ctx context.Context, resource layoutapi.ResourceKind,
) (PublishedPresentation, error) {
	return store.resolve(ctx, resource)
}

func fixtureDefinition() layoutapi.Definition {
	section := layoutapi.Section{
		ID: fixtureSectionID.String(), Order: 0,
		Translations: []layoutapi.Translation{{
			Help: "", Label: "General", Locale: "en",
		}},
		Fields: []layoutapi.Field{
			{FieldID: fixtureFieldOne.String(), Interaction: layoutapi.InteractionAuto,
				Order: 0, Widget: layoutapi.WidgetText,
				Translations: []layoutapi.Translation{}},
			{FieldID: fixtureFieldTwo.String(), Interaction: layoutapi.InteractionAuto,
				Order: 1, Widget: layoutapi.WidgetAuto,
				Translations: []layoutapi.Translation{}},
		},
	}
	profiles := systemResourceSurfaceProfiles[:5]
	surfaces := make([]layoutapi.SurfaceDefinition, len(profiles))
	for index, profile := range profiles {
		surfaces[index] = layoutapi.SurfaceDefinition{
			Kind: profile.kind, Sections: []layoutapi.Section{section},
			Translations: []layoutapi.Translation{{Locale: "en", Label: profile.label}},
		}
	}
	return layoutapi.Definition{
		SourceContract: layoutapi.DefinitionContractV2, Surfaces: surfaces,
	}
}

func fixtureRevision(
	t *testing.T,
	state layoutapi.RevisionState,
	version int64,
) layoutapi.Revision {
	t.Helper()
	definition, _, digest, err := encodeDefinition(fixtureDefinition())
	if err != nil {
		t.Fatal(err)
	}
	revision := layoutapi.Revision{
		ID: fixtureRevisionID.String(), Resource: layoutapi.ResourceUser,
		Version: version, State: state, Definition: definition,
		DefinitionSHA256: digest, CreatedAt: fixtureNow,
		UpdatedAt: fixtureNow, RowRevision: 1,
	}
	if state == layoutapi.RevisionPublished {
		revision.PublishedAt = &revision.UpdatedAt
		revision.RowRevision = 2
	}
	return revision
}

func fixtureApplication(t *testing.T, store *fixtureStore) *LayoutApplication {
	t.Helper()
	service, err := NewLayoutApplication(
		store, fixtureIDs{id: fixtureRevisionID}, fixtureClock{value: fixtureNow},
	)
	if err != nil {
		t.Fatal(err)
	}
	return service.(*LayoutApplication)
}

func fixturePresentationApplication(
	t *testing.T,
	store *fixtureStore,
) *PresentationApplication {
	t.Helper()
	service, err := NewPresentationApplication(store, fixtureClock{value: fixtureNow})
	if err != nil {
		t.Fatal(err)
	}
	return service
}

func fixtureCatalogForIDs(
	t *testing.T,
	ids ...uuid.UUID,
) presentation.FieldCatalog {
	t.Helper()
	fields := make([]presentation.Field, len(ids))
	profile, valid := presentation.DefaultValueProfile(fieldregistry.TypeText)
	if !valid {
		t.Fatal("text value profile is unavailable")
	}
	for index, id := range ids {
		field, err := presentation.NewField(
			id.String(), fieldregistry.TypeText, profile, presentation.FieldEditable,
		)
		if err != nil {
			t.Fatal(err)
		}
		fields[index] = field
	}
	catalog, err := presentation.NewFieldCatalog(fields)
	if err != nil {
		t.Fatal(err)
	}
	return catalog
}
