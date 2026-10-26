package serviceprincipal

import (
	"reflect"

	rbacports "github.com/Onurlulardan/harpy/core/internal/modules/iam/ports/rbac"
	"github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type Commands struct {
	port    CommandPort
	catalog CapabilityCatalog
	sink    rbacports.ChangeSink[transaction.TenantScope]
}

func NewCommands(
	port CommandPort,
	catalog CapabilityCatalog,
	sink rbacports.ChangeSink[transaction.TenantScope],
) (*Commands, error) {
	if commandPortNil(port) || !catalog.valid || changeSinkNil(sink) {
		return nil, ErrEvidenceMissing
	}
	return &Commands{port: port, catalog: catalog, sink: sink}, nil
}

func commandPortNil(port CommandPort) bool {
	if port == nil {
		return true
	}
	value := reflect.ValueOf(port)
	return value.Kind() == reflect.Pointer && value.IsNil()
}
