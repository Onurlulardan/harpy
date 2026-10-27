package billing

import (
	"context"
	"testing"

	platformlicenseapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/platformlicense"
	"github.com/google/uuid"
)

type platformLicenseEnvironmentStub struct {
	value PlatformLicenseHTTPEnvironment
	err   error
}

func (stub platformLicenseEnvironmentStub) Current(
	context.Context,
) (PlatformLicenseHTTPEnvironment, error) {
	return stub.value, stub.err
}

type platformLicenseAccessStub struct {
	value   PlatformLicenseHTTPAccess
	err     error
	request PlatformLicenseHTTPAccessRequest
	calls   int
}

func (stub *platformLicenseAccessStub) Verify(
	_ context.Context,
	request PlatformLicenseHTTPAccessRequest,
) (PlatformLicenseHTTPAccess, error) {
	stub.request, stub.calls = request, stub.calls+1
	return stub.value, stub.err
}

type platformLicenseApplicationStub struct {
	license        platformlicenseapp.License
	refresh        platformlicenseapp.RefreshAccepted
	operation      platformlicenseapp.RefreshOperation
	err            error
	refreshCommand platformlicenseapp.RefreshCommand
	environmentID  uuid.UUID
	operationID    uuid.UUID
}

func (stub *platformLicenseApplicationStub) Read(
	_ context.Context,
	environmentID uuid.UUID,
) (platformlicenseapp.License, error) {
	stub.environmentID = environmentID
	return stub.license, stub.err
}

func (stub *platformLicenseApplicationStub) Create(
	_ context.Context,
	command platformlicenseapp.RefreshCommand,
) (platformlicenseapp.RefreshAccepted, error) {
	stub.refreshCommand = command
	return stub.refresh, stub.err
}

func (stub *platformLicenseApplicationStub) Get(
	_ context.Context,
	environmentID uuid.UUID,
	operationID uuid.UUID,
) (platformlicenseapp.RefreshOperation, error) {
	stub.environmentID, stub.operationID = environmentID, operationID
	return stub.operation, stub.err
}

func platformLicenseHTTPV7(t *testing.T) uuid.UUID {
	t.Helper()
	value, err := uuid.NewV7()
	if err != nil {
		t.Fatal(err)
	}
	return value
}

func newPlatformLicenseHTTPFixture(
	t *testing.T,
	mode string,
) (*PlatformLicenseHTTPHandler, *platformLicenseAccessStub,
	*platformLicenseApplicationStub, PlatformLicenseHTTPEnvironment) {
	t.Helper()
	environment := PlatformLicenseHTTPEnvironment{
		EnvironmentID:  platformLicenseHTTPV7(t).String(),
		TenantID:       "00000000-0000-4000-8000-000000000002",
		DeploymentMode: mode,
	}
	access := &platformLicenseAccessStub{value: PlatformLicenseHTTPAccess{
		ActorID:      "00000000-0000-4000-8000-000000000001",
		TenantID:     environment.TenantID,
		SessionID:    "00000000-0000-4000-8000-000000000003",
		TokenVersion: 1, AuthorizationVersion: 2, PlatformClock: 3,
	}}
	application := &platformLicenseApplicationStub{}
	handler, err := NewPlatformLicenseHTTPHandler(
		PlatformLicenseHTTPDependencies{
			Access:      access,
			Environment: platformLicenseEnvironmentStub{value: environment},
			Application: application, Report: func(error) {},
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	return handler, access, application, environment
}
