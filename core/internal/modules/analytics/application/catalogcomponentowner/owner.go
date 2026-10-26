package catalogcomponentowner

import (
	"fmt"
	"reflect"

	contract "github.com/Onurlulardan/harpy/contracts/catalogcomponent/v1"
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
	lifecycle LifecycleStoreBinder
}

func New(base BaseOwner, lifecycle LifecycleStoreBinder) (*Owner, error) {
	if nilPort(base) || nilPort(lifecycle) {
		return nil, fmt.Errorf(
			"Analytics Catalog owner dependencies are invalid",
		)
	}
	return &Owner{BaseOwner: base, lifecycle: lifecycle}, nil
}

func nilPort(value any) bool {
	if value == nil {
		return true
	}
	seen := reflect.ValueOf(value)
	return (seen.Kind() == reflect.Pointer ||
		seen.Kind() == reflect.Interface) && seen.IsNil()
}
