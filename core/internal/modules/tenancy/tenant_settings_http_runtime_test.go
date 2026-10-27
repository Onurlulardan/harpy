package tenancy

import (
	"context"
	"errors"
	"reflect"
	"testing"

	settings "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/application/tenantsettings"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

var errSettingsScopeReached = errors.New("tenant settings scope reached")

type settingsAuthorityFake struct {
	value    TenantSettingsAuthority
	tenantID uuid.UUID
}

func (fake *settingsAuthorityFake) ResolveTenantSettingsAuthority(_ context.Context,
	tenantID uuid.UUID,
) (TenantSettingsAuthority, error) {
	fake.tenantID = tenantID
	return fake.value, nil
}

type settingsCoordinatorFake struct{ calls int }

func (fake *settingsCoordinatorFake) InTenant(context.Context, platformtx.ScopeRequest,
	func(context.Context, platformtx.TenantScope) error,
) error {
	fake.calls++
	return errSettingsScopeReached
}

type settingsAdmitFake struct {
	calls        int
	groups       []platformtx.CommandAuthorityGroup
	domains      []platformtx.CommandAuthorityDomain
	participants []string
}

type settingsSnapshotIssueFake struct {
	calls       int
	authorities map[string]int64
}

func (fake *settingsSnapshotIssueFake) issue(_ string,
	authorities map[string]int64,
) (platformtx.VerifiedOperation, error) {
	fake.calls++
	fake.authorities = make(map[string]int64, len(authorities))
	for owner, generation := range authorities {
		fake.authorities[owner] = generation
	}
	return platformtx.VerifiedOperation{}, nil
}

func (fake *settingsAdmitFake) admit(_ uuid.UUID, _ uuid.UUID, _ string,
	groups []platformtx.CommandAuthorityGroup,
	domains []platformtx.CommandAuthorityDomain,
	participants []string,
) (platformtx.ScopeRequest, error) {
	fake.calls++
	fake.groups = append([]platformtx.CommandAuthorityGroup(nil), groups...)
	fake.domains = append([]platformtx.CommandAuthorityDomain(nil), domains...)
	fake.participants = append([]string(nil), participants...)
	return platformtx.ScopeRequest{}, nil
}

func TestTenantSettingsRuntimeBuildsMinimalNameCommandAuthority(t *testing.T) {
	resolver := &settingsAuthorityFake{value: TenantSettingsAuthority{
		IAMGroup: 4, IAM: 5, SystemResourceLayout: 6}}
	coordinator := &settingsCoordinatorFake{}
	admitter := &settingsAdmitFake{}
	runtime := &TenantSettingsHTTPRuntime{resolver: resolver, coordinator: coordinator,
		admitCommand: admitter.admit}
	access := settingsRuntimeAccess(t, TenantSettingsPermissionManage,
		TenantSettingsOperationUpdate)
	name := "Updated"
	_, err := runtime.Update(t.Context(), access, settings.UpdateCommand{
		Name: &name, ExpectedRevision: 1, IdempotencyKey: "tenant-settings-0001"})
	if !errors.Is(err, settings.ErrUnavailable) || coordinator.calls != 0 || admitter.calls != 0 {
		t.Fatalf("missing tenancy authority err=%v coordinator=%d admit=%d", err,
			coordinator.calls, admitter.calls)
	}
	resolver.value.TenancyGroup, resolver.value.Tenancy = 2, 3
	_, err = runtime.Update(t.Context(), access, settings.UpdateCommand{
		Name: &name, ExpectedRevision: 1, IdempotencyKey: "tenant-settings-0001"})
	if !errors.Is(err, errSettingsScopeReached) || coordinator.calls != 1 ||
		resolver.tenantID != settingsTestTenant || admitter.calls != 1 {
		t.Fatalf("name command err=%v coordinator=%d admit=%d tenant=%s", err,
			coordinator.calls, admitter.calls, resolver.tenantID)
	}
	assertSettingsCommandAuthority(t, admitter, false)
}

func TestTenantSettingsRuntimeRequiresNotificationsForSMTPCommands(t *testing.T) {
	resolver := &settingsAuthorityFake{value: TenantSettingsAuthority{
		TenancyGroup: 2, Tenancy: 3, IAMGroup: 4, IAM: 5,
		SystemResourceLayout: 6}}
	coordinator := &settingsCoordinatorFake{}
	admitter := &settingsAdmitFake{}
	runtime := &TenantSettingsHTTPRuntime{resolver: resolver, coordinator: coordinator,
		admitCommand: admitter.admit}
	access := settingsRuntimeAccess(t, TenantSettingsPermissionManage,
		TenantSettingsOperationUpdate)
	_, err := runtime.Update(t.Context(), access, settings.UpdateCommand{
		SMTP: &settings.SMTPMutation{}, ExpectedRevision: 1,
		IdempotencyKey: "tenant-settings-0001"})
	if !errors.Is(err, settings.ErrUnavailable) || coordinator.calls != 0 {
		t.Fatalf("SMTP authority err=%v calls=%d", err, coordinator.calls)
	}
	resolver.value.NotificationsGroup, resolver.value.Notifications = 7, 8
	_, err = runtime.Update(t.Context(), access, settings.UpdateCommand{
		SMTP: &settings.SMTPMutation{}, ExpectedRevision: 1,
		IdempotencyKey: "tenant-settings-0001"})
	if !errors.Is(err, errSettingsScopeReached) || coordinator.calls != 1 || admitter.calls != 1 {
		t.Fatalf("SMTP command err=%v coordinator=%d admit=%d", err,
			coordinator.calls, admitter.calls)
	}
	assertSettingsCommandAuthority(t, admitter, true)
}

func TestTenantSettingsRuntimeBuildsCrossOwnerSnapshotAuthority(t *testing.T) {
	resolver := &settingsAuthorityFake{value: TenantSettingsAuthority{
		Tenancy: 2, IAM: 3, Notifications: 4, Billing: 5}}
	coordinator := &settingsCoordinatorFake{}
	issuer := &settingsSnapshotIssueFake{}
	runtime := &TenantSettingsHTTPRuntime{resolver: resolver, coordinator: coordinator,
		issueSnapshot: issuer.issue}
	access := settingsRuntimeAccess(t, TenantSettingsPermissionRead,
		TenantSettingsOperationGet)
	_, err := runtime.Get(t.Context(), access)
	if !errors.Is(err, errSettingsScopeReached) || coordinator.calls != 1 || issuer.calls != 1 {
		t.Fatalf("snapshot err=%v coordinator=%d issue=%d", err,
			coordinator.calls, issuer.calls)
	}
	want := map[string]int64{
		"billing": 5, "iam": 3, "notifications": 4, "tenancy": 2,
	}
	if !reflect.DeepEqual(issuer.authorities, want) {
		t.Fatalf("snapshot authorities=%v want=%v", issuer.authorities, want)
	}
}

func assertSettingsCommandAuthority(t *testing.T, actual *settingsAdmitFake,
	includeNotifications bool,
) {
	t.Helper()
	groups := []platformtx.CommandAuthorityGroup{
		{ID: "iam_system_resources", Generation: 4},
		{ID: "tenancy", Generation: 2},
	}
	domains := []platformtx.CommandAuthorityDomain{
		{Name: "iam", Generation: 5},
		{Name: "system_resource_layout", Generation: 6},
		{Name: "tenancy", Generation: 3},
	}
	participants := []string{"tenancy"}
	if includeNotifications {
		groups = append(groups,
			platformtx.CommandAuthorityGroup{ID: "notifications", Generation: 7})
		domains = append(domains,
			platformtx.CommandAuthorityDomain{Name: "notifications", Generation: 8})
		participants = append(participants, "notifications")
	}
	if !reflect.DeepEqual(actual.groups, groups) ||
		!reflect.DeepEqual(actual.domains, domains) ||
		!reflect.DeepEqual(actual.participants, participants) {
		t.Fatalf("authority groups=%+v domains=%+v participants=%v",
			actual.groups, actual.domains, actual.participants)
	}
}

func settingsRuntimeAccess(t *testing.T, permission TenantSettingsPermission,
	operation string,
) TenantSettingsAccess {
	t.Helper()
	access, err := NewTenantSettingsAccess(settingsTestPrincipal, settingsTestTenant,
		1, 2, 3, permission, operation)
	if err != nil {
		t.Fatal(err)
	}
	return access
}
