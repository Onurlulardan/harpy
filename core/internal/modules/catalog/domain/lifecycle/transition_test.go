package lifecycle

import (
	"errors"
	"testing"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/catalog/domain/packagecontract"
)

const (
	lifecyclePackage   = "018f1000-0000-7000-8000-000000000001"
	lifecycleInstall   = "018f1000-0000-7000-8000-000000000002"
	lifecycleOperation = "018f1000-0000-7000-8000-000000000003"
	lifecycleComponent = "018f1000-0000-7000-8000-000000000004"
)

func TestInstallationTransitionsPreserveEpochAndVersion(t *testing.T) {
	installation := installingFixture(t)
	version := versionFixture(t, "1.0.0")
	active, err := installation.Activate(version)
	if err != nil || active.ReferenceEpoch != 1 || active.RowVersion != 2 {
		t.Fatalf("activate = %#v, %v", active, err)
	}
	disabled, err := active.Disable()
	if err != nil || disabled.ActiveVersion.Compare(version) != 0 {
		t.Fatalf("disable = %#v, %v", disabled, err)
	}
	retired, err := disabled.Retire()
	if err != nil || retired.ReferenceEpoch != 2 {
		t.Fatalf("retire = %#v, %v", retired, err)
	}
	pending, err := retired.BeginPurge()
	if err != nil || pending.ReferenceEpoch != 3 {
		t.Fatalf("begin purge = %#v, %v", pending, err)
	}
	restored, err := pending.CancelPurge()
	if err != nil || restored.State != InstallationRetired || restored.ReferenceEpoch != 4 {
		t.Fatalf("cancel purge = %#v, %v", restored, err)
	}
}

func TestOperationCancellationClosesBeforeDestructiveBoundary(t *testing.T) {
	operation := operationFixture(t, OperationPurge)
	applying, _ := operation.Advance(OperationApplying)
	destructive, err := applying.StartDestructive(canonicalNow())
	if err != nil {
		t.Fatalf("start destructive: %v", err)
	}
	if _, err = destructive.RequestCancel(); !errors.Is(err, ErrCancellationLate) {
		t.Fatalf("late cancellation error = %v", err)
	}
	install := operationFixture(t, OperationInstall)
	cancelled, err := install.RequestCancel()
	if err != nil || cancelled.State != OperationCancelled {
		t.Fatalf("planned cancel = %#v, %v", cancelled, err)
	}
}

func TestDemoOperationsRejectCancellationAfterApplyStarts(t *testing.T) {
	for _, kind := range []OperationKind{OperationDemoInstall, OperationDemoRemove} {
		operation := operationFixture(t, kind)
		cancelled, err := operation.RequestCancel()
		if err != nil || cancelled.State != OperationCancelled {
			t.Fatalf("%s planned cancel = %#v, %v", kind, cancelled, err)
		}
		applying, err := operation.Advance(OperationApplying)
		if err != nil {
			t.Fatal(err)
		}
		if _, err = applying.RequestCancel(); !errors.Is(err, ErrCancellationLate) {
			t.Fatalf("%s applying cancellation error = %v", kind, err)
		}
	}
}

func TestFailedInstallationCanRetryOnlyFromFailed(t *testing.T) {
	installing := installingFixture(t)
	if _, err := installing.Retry(); !errors.Is(err, ErrTransition) {
		t.Fatalf("installing retry error = %v", err)
	}
	failed, err := installing.Fail()
	if err != nil {
		t.Fatal(err)
	}
	retrying, err := failed.Retry()
	if err != nil || retrying.State != InstallationInstalling || retrying.RowVersion != 3 {
		t.Fatalf("retry = %#v, %v", retrying, err)
	}
}

func installingFixture(t *testing.T) Installation {
	t.Helper()
	id, _ := packagecontract.ParseStableID(lifecycleInstall)
	packageID, _ := packagecontract.ParsePackageID(lifecyclePackage)
	value, err := NewInstalling(id, packageID)
	if err != nil {
		t.Fatal(err)
	}
	return value
}

func operationFixture(t *testing.T, kind OperationKind) Operation {
	t.Helper()
	packageID, _ := packagecontract.ParsePackageID(lifecyclePackage)
	componentID, _ := packagecontract.ParseComponentID(lifecycleComponent)
	plan, err := NewPlan(packageID, versionFixture(t, "1.0.0"), nil, []ComponentPin{{
		ComponentID: componentID, Owner: "catalog", OwnerGeneration: 1,
		Contract: "catalog.v2", ContentSHA256: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
	}})
	if err != nil {
		t.Fatal(err)
	}
	id, _ := packagecontract.ParseStableID(lifecycleOperation)
	value, err := NewOperation(id, kind, plan, canonicalNow())
	if err != nil {
		t.Fatal(err)
	}
	return value
}

func versionFixture(t *testing.T, raw string) packagecontract.Version {
	t.Helper()
	value, err := packagecontract.ParseVersion(raw)
	if err != nil {
		t.Fatal(err)
	}
	return value
}

func canonicalNow() time.Time { return time.Date(2026, 7, 18, 12, 0, 0, 0, time.UTC) }
