package authorization

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
)

func TestMandatoryFilterAccessorsDefensivelyClone(t *testing.T) {
	left := mandatoryLeaf(t)
	right := mandatoryLeaf(t)
	filter, err := NewMandatoryOr([]MandatoryFilter{left, right})
	if err != nil {
		t.Fatal(err)
	}
	kind, kindErr := filter.Kind()
	predicate, found, predicateErr := filter.Predicate()
	children, childrenErr := filter.Children()
	if kindErr != nil || kind != api.RequestFilterOr || predicateErr != nil || found ||
		predicate.Validate() == nil || childrenErr != nil || len(children) != 2 {
		t.Fatal("mandatory filter accessors are invalid")
	}
	children[0] = MandatoryFilter{}
	if filter.Validate() != nil {
		t.Fatal("returned children mutated mandatory filter")
	}
	if fmt.Sprintf("%v", filter) != authorizationRedaction ||
		filter.LogValue().String() != authorizationRedaction {
		t.Fatal("mandatory filter was not redacted")
	}
	if _, err := json.Marshal(filter); err == nil {
		t.Fatal("mandatory filter allowed direct JSON marshal")
	}
}
