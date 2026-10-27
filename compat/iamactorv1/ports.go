package iamactorv1

import (
	"context"
	"reflect"

	actorassertionv1 "github.com/Onurlulardan/harpy/contracts/actorassertion/v1"
)

type Row interface {
	Scan(...any) error
}

// Transaction is deliberately narrower than a connection. The V1 edge owns
// begin/commit/rollback and must pass the same harpy_v1_runtime transaction
// to Consume and its business effect.
type Transaction interface {
	QueryRow(context.Context, string, ...any) Row
}

type Effect func(
	context.Context,
	Transaction,
	actorassertionv1.Claims,
) error

func nilTransaction(transaction Transaction) bool {
	if transaction == nil {
		return true
	}
	value := reflect.ValueOf(transaction)
	switch value.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map,
		reflect.Pointer, reflect.Slice:
		return value.IsNil()
	default:
		return false
	}
}
