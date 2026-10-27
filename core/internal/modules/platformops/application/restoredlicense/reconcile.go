package restoredlicense

import (
	"context"
	"fmt"
	"regexp"
	"time"

	restoredlicensev1 "github.com/Onurlulardan/harpy/contracts/restoredlicense/v1"
	capability "github.com/Onurlulardan/harpy/core/internal/capabilities/restoredlicense/api"
	"github.com/google/uuid"
)

var digestPattern = regexp.MustCompile(`^[0-9a-f]{64}$`)

type Grant = capability.Grant
type Scope = capability.Scope
type Binding = capability.Binding

type Input struct {
	OperationID            uuid.UUID
	EnvironmentID          uuid.UUID
	RestoreReceiptID       uuid.UUID
	RestoreReceiptSHA256   string
	LicenseID              uuid.UUID
	LicenseeOrganizationID uuid.UUID
	DeliveryKind           string
	Issuer                 string
	PreviousSequence       int64
	PreviousSHA256         string
	NextSequence           int64
	NextSHA256             string
	AuthorityRevision      int64
}

type Result struct {
	Grant       Grant
	DatabaseNow time.Time
}

type Store interface {
	ReconcileRestoredLicense(context.Context, Input) (Result, error)
}

func Reconcile(
	ctx context.Context,
	store Store,
	input Input,
) (Result, error) {
	if ctx == nil || ctx.Err() != nil || store == nil || !validInput(input) {
		return Result{}, fmt.Errorf("restored license grant input is invalid")
	}
	result, err := store.ReconcileRestoredLicense(ctx, input)
	if err != nil {
		return Result{}, fmt.Errorf("mint restored license grant: %w", err)
	}
	expected := capability.Binding{
		OperationID: input.OperationID, EnvironmentID: input.EnvironmentID,
		RestoreReceiptID:       input.RestoreReceiptID,
		RestoreReceiptSHA256:   input.RestoreReceiptSHA256,
		LicenseID:              input.LicenseID,
		LicenseeOrganizationID: input.LicenseeOrganizationID,
		DeliveryKind:           input.DeliveryKind, Issuer: input.Issuer,
		PreviousSequence: input.PreviousSequence,
		PreviousSHA256:   input.PreviousSHA256,
		NextSequence:     input.NextSequence, NextSHA256: input.NextSHA256,
		ManifestSHA256: restoredlicensev1.ManifestSHA256(),
	}
	actual := result.Grant.Binding
	expected.LicenseBindingID = actual.LicenseBindingID
	if actual != expected || actual.LicenseBindingID == uuid.Nil ||
		result.Grant.Revision != input.AuthorityRevision ||
		!validClock(result.DatabaseNow) || !validClock(result.Grant.LeaseUntil) ||
		!result.Grant.LeaseUntil.After(result.DatabaseNow) {
		return Result{}, fmt.Errorf("restored license grant response is invalid")
	}
	return result, nil
}

func Run(
	grant Grant,
	databaseNow time.Time,
	callback func(Scope) error,
) error {
	return capability.Run(grant, databaseNow, callback)
}

func validInput(input Input) bool {
	return input.OperationID != uuid.Nil &&
		input.OperationID.Version() == uuid.Version(7) &&
		input.EnvironmentID != uuid.Nil && input.RestoreReceiptID != uuid.Nil &&
		input.LicenseID != uuid.Nil && input.LicenseeOrganizationID != uuid.Nil &&
		digestPattern.MatchString(input.RestoreReceiptSHA256) &&
		(input.DeliveryKind == "connected" || input.DeliveryKind == "airgap" ||
			input.DeliveryKind == "development") &&
		input.Issuer != "" && len(input.Issuer) <= 256 &&
		input.PreviousSequence > 0 && input.NextSequence > input.PreviousSequence &&
		digestPattern.MatchString(input.PreviousSHA256) &&
		digestPattern.MatchString(input.NextSHA256) && input.AuthorityRevision > 0
}

func validClock(value time.Time) bool {
	return !value.IsZero() && value.Location() == time.UTC
}
