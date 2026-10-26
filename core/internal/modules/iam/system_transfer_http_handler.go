package iam

import (
	"fmt"
	"reflect"
)

type SystemTransferHTTPErrorSink func(error)

type SystemTransferHTTPDependencies struct {
	Access      SystemTransferAccessVerifier
	Application SystemTransferApplication
	Report      SystemTransferHTTPErrorSink
}

type SystemTransferHTTP struct {
	access      SystemTransferAccessVerifier
	application SystemTransferApplication
	report      SystemTransferHTTPErrorSink
}

func NewSystemTransferHTTP(dependencies SystemTransferHTTPDependencies) (
	*SystemTransferHTTP, error) {
	if systemTransferNil(dependencies.Access) ||
		systemTransferNil(dependencies.Application) || dependencies.Report == nil {
		return nil, fmt.Errorf("IAM transfer HTTP dependencies are required")
	}
	return &SystemTransferHTTP{access: dependencies.Access,
		application: dependencies.Application, report: dependencies.Report}, nil
}

func (handler *SystemTransferHTTP) Ready() bool {
	return handler != nil && !systemTransferNil(handler.access) &&
		!systemTransferNil(handler.application) && handler.report != nil
}

func systemTransferNil(value any) bool {
	if value == nil {
		return true
	}
	seen := reflect.ValueOf(value)
	switch seen.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map,
		reflect.Pointer, reflect.Slice:
		return seen.IsNil()
	default:
		return false
	}
}
