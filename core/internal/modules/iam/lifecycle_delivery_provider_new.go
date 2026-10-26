package iam

import (
	"reflect"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/platform/async/effectexecution"
)

func newLifecycleDeliveryProvider(state lifecycleDeliveryState,
	opener lifecycleCredentialOpener, sender lifecycleSMTPSender,
	now func() time.Time) (*lifecycleDeliveryProvider, error) {
	identity, err := effectexecution.NewIdentity("iam.smtp", "send_lifecycle_credential")
	if err != nil || missingLifecycleDependency(state) ||
		missingLifecycleDependency(opener) || missingLifecycleDependency(sender) || now == nil {
		return nil, ErrLifecycleDeliveryRuntime
	}
	return &lifecycleDeliveryProvider{identity: identity, state: state,
		opener: opener, smtp: sender, now: now}, nil
}

func missingLifecycleDependency(value any) bool {
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
