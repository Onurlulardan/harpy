package postgres

import (
	"context"
	"fmt"
	"reflect"

	"github.com/Onurlulardan/harpy/core/internal/platform/contractretirement"
	"github.com/jackc/pgx/v5"
)

type ContractRetirementQueryer interface {
	Query(context.Context, string, ...any) (pgx.Rows, error)
	QueryRow(context.Context, string, ...any) pgx.Row
}

type ContractRetirementDatabase struct {
	queryer ContractRetirementQueryer
}

func NewContractRetirementDatabase(
	queryer ContractRetirementQueryer,
) (*ContractRetirementDatabase, error) {
	if queryer == nil || isNilContractRetirementQueryer(queryer) {
		return nil, fmt.Errorf("contract retirement PostgreSQL queryer is required")
	}
	return &ContractRetirementDatabase{queryer: queryer}, nil
}

func isNilContractRetirementQueryer(queryer ContractRetirementQueryer) bool {
	value := reflect.ValueOf(queryer)
	switch value.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map,
		reflect.Pointer, reflect.Slice:
		return value.IsNil()
	default:
		return false
	}
}

var _ contractretirement.RuntimePort = (*ContractRetirementDatabase)(nil)
var _ contractretirement.AuthorityPort = (*ContractRetirementDatabase)(nil)
