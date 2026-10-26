package systemtransferartifact

import (
	"reflect"
	"time"

	fcartv1 "github.com/Onurlulardan/harpy/contracts/fcart/v1"
	artifactobject "github.com/Onurlulardan/harpy/core/internal/capabilities/artifactobject/api"
	spoolapi "github.com/Onurlulardan/harpy/core/internal/capabilities/artifactspool/api"
)

type Clock interface{ Now() time.Time }

type ObjectStore = artifactobject.Store
type SpoolManager = spoolapi.Manager

type Runtime struct {
	imports, exports ObjectStore
	importKeys       *fcartv1.EnvelopeKeyring
	exportKeys       *fcartv1.EnvelopeKeyring
	spools           *SpoolManager
	clock            Clock
}

func NewRuntime(imports, exports ObjectStore,
	importKeys, exportKeys *fcartv1.EnvelopeKeyring,
	spools *SpoolManager, clock Clock) (*Runtime, error) {
	dependencies := []any{imports, exports, importKeys, exportKeys, spools, clock}
	for _, dependency := range dependencies {
		if nilDependency(dependency) {
			return nil, errUnavailable
		}
	}
	return &Runtime{imports: imports, exports: exports, importKeys: importKeys,
		exportKeys: exportKeys, spools: spools, clock: clock}, nil
}

func nilDependency(value any) bool {
	if value == nil {
		return true
	}
	reflected := reflect.ValueOf(value)
	switch reflected.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map,
		reflect.Pointer, reflect.Slice:
		return reflected.IsNil()
	default:
		return false
	}
}
