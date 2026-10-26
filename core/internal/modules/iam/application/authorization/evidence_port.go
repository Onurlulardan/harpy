package authorization

import (
	"context"
	"reflect"

	"github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type EvidencePort interface {
	Authorize(context.Context, transaction.TenantScope, Request) (Snapshot, error)
}

func evidencePortIsNil(port EvidencePort) bool {
	if port == nil {
		return true
	}
	value := reflect.ValueOf(port)
	switch value.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map,
		reflect.Pointer, reflect.Slice:
		return value.IsNil()
	default:
		return false
	}
}
