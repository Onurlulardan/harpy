package platformlicense

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
)

type storeFixture struct {
	license       License
	refresh       RefreshAccepted
	operation     RefreshOperation
	err           error
	refreshInput  RefreshCommand
	environmentID uuid.UUID
	operationID   uuid.UUID
}

func (fixture *storeFixture) ReadLicense(
	_ context.Context,
	environmentID uuid.UUID,
) (License, error) {
	fixture.environmentID = environmentID
	return fixture.license, fixture.err
}

func (fixture *storeFixture) Create(
	_ context.Context,
	command RefreshCommand,
) (RefreshAccepted, error) {
	fixture.refreshInput = command
	return fixture.refresh, fixture.err
}

func (fixture *storeFixture) Get(
	_ context.Context,
	environmentID uuid.UUID,
	operationID uuid.UUID,
) (RefreshOperation, error) {
	fixture.environmentID = environmentID
	fixture.operationID = operationID
	return fixture.operation, fixture.err
}

func testV7(t *testing.T) uuid.UUID {
	t.Helper()
	value, err := uuid.NewV7()
	if err != nil {
		t.Fatal(err)
	}
	return value
}

func validLicenseFixture(t *testing.T, environmentID uuid.UUID) License {
	t.Helper()
	return License{
		LicenseID: testV7(t), EnvironmentID: environmentID,
		SnapshotSequence: 7,
		SnapshotSHA256:   "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		State:            "valid", ReleaseChannel: "stable",
		Features: map[string]bool{"workflow": true},
		Limits:   map[string]Limit{"users": {Value: 100}},
		DatabaseNow: time.Date(
			2026, 7, 23, 12, 0, 0, 0, time.UTC,
		),
	}
}
