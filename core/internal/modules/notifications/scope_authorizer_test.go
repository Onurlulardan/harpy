package notifications

import (
	"context"
	"errors"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/application"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func TestIntentAuthorizationRejectsCrossProducerNotificationType(t *testing.T) {
	err := (scopeAuthorizer{}).AuthorizeIntent(context.Background(), nil,
		"notifications", "iam.password_reset")
	if !errors.Is(err, application.ErrForbidden) {
		t.Fatalf("cross-producer notification type error = %v", err)
	}
}

func TestSMTPSettingsRejectClosedAuthorityScope(t *testing.T) {
	err := authorizeNotificationScope(t.Context(), platformtx.TenantScope{},
		application.PermissionSMTPSettingsWrite)
	if !errors.Is(err, application.ErrForbidden) {
		t.Fatalf("closed SMTP settings authority error = %v", err)
	}
}

func TestSMTPSettingsOperationsDoNotBroadenGenericSend(t *testing.T) {
	operations := notificationPermissionOperations[application.PermissionSend]
	for _, operation := range []string{"tenancy.settings.get",
		"tenancy.settings.update", "tenancy.settings.test_email"} {
		if _, found := operations[operation]; found {
			t.Fatalf("generic send unexpectedly allows %s", operation)
		}
	}
	if _, found := notificationPermissionOperations[application.PermissionSMTPSettingsWrite]["tenancy.settings.update"]; !found {
		t.Fatal("exact tenant SMTP update operation is not authorized")
	}
}
