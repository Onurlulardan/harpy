package aiworkspace

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	catalogai "github.com/Onurlulardan/harpy/core/internal/modules/catalog/api/aiworkspace"
	"github.com/google/uuid"
)

func TestUnavailableOwnerIsNeverReportedAsInvalidInput(t *testing.T) {
	owner := (*Owner)(nil)
	checks := []func() error{
		func() error {
			_, err := owner.List(context.Background(), catalogai.ListQuery{Limit: 20})
			return err
		},
		func() error {
			_, err := owner.Search(context.Background(),
				catalogai.SearchQuery{Query: "customer", Limit: 10})
			return err
		},
		func() error {
			_, err := owner.Describe(context.Background(),
				"forms/0198eac5-5ee9-7e0f-8c4a-8493a072f851")
			return err
		},
		func() error {
			_, err := owner.Recent(context.Background(),
				catalogai.RecentQuery{Since: time.Hour, Limit: 10})
			return err
		},
		func() error {
			_, err := owner.History(context.Background(),
				"forms/0198eac5-5ee9-7e0f-8c4a-8493a072f851", 10)
			return err
		},
	}
	for index, check := range checks {
		if err := check(); !errors.Is(err, catalogai.ErrUnavailable) {
			t.Fatalf("check %d error=%v", index, err)
		}
	}
}

func TestMenuProjectionUsesStableTargetPathAndCatalogResource(t *testing.T) {
	itemID, targetID := uuid.Must(uuid.NewV7()), uuid.Must(uuid.NewV7())
	projection, valid := (menuItem{ItemID: itemID, TargetID: &targetID,
		Kind: "FORM", Label: "Customers", UpdatedAt: time.Now().UTC()}).projection()
	if !valid || projection.Path != "forms/"+targetID.String() ||
		projection.Slug != targetID.String() || len(projection.Requirements) != 1 ||
		projection.Requirements[0].Resource != "catalog/menu/"+itemID.String() ||
		projection.Requirements[0].Permission != "ai.workspace.read" {
		t.Fatalf("projection=%+v valid=%t", projection, valid)
	}
}

func TestMenuDetailExposesOnlyItsAuthoritativeTargetReference(t *testing.T) {
	targetID := uuid.Must(uuid.NewV7())
	reference, valid := (menuItem{ItemID: uuid.Must(uuid.NewV7()), TargetID: &targetID,
		Kind: "FORM"}).targetReference()
	if !valid || !reference.Valid() || reference.Kind != "FORM" || reference.ID != targetID {
		t.Fatalf("target reference=%+v valid=%t", reference, valid)
	}
}

func TestHistoryQueryFollowsDefinitionInsteadOfVersionLocalItemID(t *testing.T) {
	if !strings.Contains(historySQL, "selected_definition") ||
		!strings.Contains(historySQL, "version.definition_id=(SELECT definition_id") ||
		strings.Contains(historySQL, "WHERE item.item_id=$1\nORDER BY") {
		t.Fatalf("history query does not follow the immutable definition: %s", historySQL)
	}
}
