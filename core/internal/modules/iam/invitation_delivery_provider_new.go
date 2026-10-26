package iam

import (
	"reflect"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/platform/async/effectexecution"
)

func newInvitationDeliveryProvider(
	state invitationDeliveryState,
	opener invitationCredentialOpener,
	sender invitationSMTPSender,
	now func() time.Time,
) (*invitationDeliveryProvider, error) {
	identity, err := effectexecution.NewIdentity("iam.smtp", "send_invitation")
	if err != nil || missingInvitationDependency(state) ||
		missingInvitationDependency(opener) || missingInvitationDependency(sender) || now == nil {
		return nil, ErrInvitationDeliveryRuntime
	}
	return &invitationDeliveryProvider{
		identity: identity, state: state, opener: opener, smtp: sender, now: now,
	}, nil
}

func missingInvitationDependency(value any) bool {
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
