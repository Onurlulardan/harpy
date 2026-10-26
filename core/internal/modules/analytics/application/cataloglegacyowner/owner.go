package cataloglegacyowner

import (
	"fmt"
	"reflect"

	contract "github.com/Onurlulardan/harpy/contracts/catalogcomponent/v1"
	analyticsforms "github.com/Onurlulardan/harpy/core/internal/modules/analytics/api/formsreference"
	"github.com/Onurlulardan/harpy/core/internal/modules/catalog/api/componentowner"
)

type BaseOwner interface {
	componentowner.Owner
	componentowner.Provisioner
	FormsDependencySlugs(
		contract.ProvisionRequest,
	) ([]string, bool, error)
}

type Owner struct {
	BaseOwner
	references analyticsforms.Coordinator
}

func New(
	base BaseOwner,
	references analyticsforms.Coordinator,
) (*Owner, error) {
	if nilPort(base) || nilPort(references) {
		return nil, fmt.Errorf(
			"Analytics legacy Catalog owner dependencies are invalid",
		)
	}
	return &Owner{BaseOwner: base, references: references}, nil
}

func nilPort(value any) bool {
	if value == nil {
		return true
	}
	seen := reflect.ValueOf(value)
	return (seen.Kind() == reflect.Pointer ||
		seen.Kind() == reflect.Interface) && seen.IsNil()
}
