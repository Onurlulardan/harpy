package profilepostgres

import (
	"errors"
	"testing"

	profileapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/profile"
	identitydomain "github.com/Onurlulardan/harpy/core/internal/modules/iam/domain/identity"
	"github.com/google/uuid"
)

func TestProfileFieldIDsResolvePinnedHistoricalSnapshot(t *testing.T) {
	historicalID := uuid.MustParse("43000000-0000-4000-8000-000000000041")
	field, err := profileapp.NewField(profileapp.FieldInput{
		ID: historicalID, Resource: profileapp.ResourceOrganization,
		Key: "display_name", Type: profileapp.LogicalString,
		Cardinality: profileapp.CardinalitySingle, Required: true,
		Classification: profileapp.ClassificationPublic,
		Normalization:  profileapp.NormalizationNoneV1,
		State:          profileapp.FieldActive,
		Validation:     profileapp.ValidationInput{MaximumBytes: 128},
	})
	if err != nil {
		t.Fatal(err)
	}
	schema, err := profileapp.NewPublishedSchema(profileapp.SchemaInput{
		ID:       uuid.MustParse("41000000-0000-4000-8000-000000000041"),
		Resource: profileapp.ResourceOrganization, Version: 1,
		State: profileapp.SchemaRetired, Fields: []profileapp.Field{field},
	})
	if err != nil {
		t.Fatal(err)
	}
	aliases, err := profileFieldIDs(schema, []string{"display_name"})
	if err != nil ||
		aliases["display_name"] != identitydomain.IdentityID(historicalID) {
		t.Fatalf("historical aliases = %#v/%v", aliases, err)
	}
}

func TestProfileFieldIDsRejectMissingOrRetiredAlias(t *testing.T) {
	field := profileReadField(
		t, uuid.MustParse("43000000-0000-4000-8000-000000000051"), 128,
	)
	schema := profileReadSchema(
		t, "41000000-0000-4000-8000-000000000051", 1,
		profileapp.SchemaRetired, field,
	)
	if _, err := profileFieldIDs(schema, []string{"missing"}); !errors.Is(
		err, identitydomain.ErrInvalidCommand,
	) {
		t.Fatalf("missing alias error = %v", err)
	}
}
