package offering

import (
	"crypto/sha256"
	"testing"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
)

func TestPlatformVisibilityIsAllAndDeniedVisibilityIsEmpty(t *testing.T) {
	resource := sha256.Sum256([]byte("resource"))
	platform, err := policyVisibilities(resource, false, true, zeroGlobalRequest())
	if err != nil || len(platform) != 1 {
		t.Fatalf("platform visibility = %d, %v", len(platform), err)
	}
	denied, err := policyVisibilities(resource, false, false, zeroGlobalRequest())
	if err != nil || len(denied) != 0 {
		t.Fatalf("denied visibility = %d, %v", len(denied), err)
	}
}

func TestPublicVisibilityRequiresAvailableLatestOfferingAtSnapshot(t *testing.T) {
	instant := time.Date(2026, 7, 17, 15, 4, 5, 0, time.UTC)
	filter, err := publicFilter(instant)
	if err != nil {
		t.Fatal(err)
	}
	kind, err := filter.Kind()
	children, childrenErr := filter.Children()
	if err != nil || childrenErr != nil || kind != api.RequestFilterAnd || len(children) != 6 {
		t.Fatalf("public root shape = %v, %d", kind, len(children))
	}
	wanted := map[string]api.Operator{
		fieldLifecycle: api.OperatorEq, fieldPublicationState: api.OperatorEq,
		fieldPublic: api.OperatorEq, fieldAvailable: api.OperatorEq,
		fieldEffectiveStart: api.OperatorLTE,
	}
	for _, child := range children[:5] {
		predicate, found, err := child.Predicate()
		if err != nil || !found {
			t.Fatal("public predicate missing")
		}
		field, _ := predicate.FieldID()
		operator, _ := predicate.Operator()
		if wanted[field.String()] != operator {
			t.Fatalf("predicate %s operator = %v", field, operator)
		}
		delete(wanted, field.String())
	}
	endingKind, _ := children[5].Kind()
	ending, _ := children[5].Children()
	if len(wanted) != 0 || endingKind != api.RequestFilterOr || len(ending) != 2 {
		t.Fatal("effective_end public boundary is incomplete")
	}
}
