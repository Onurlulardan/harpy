package licenseauthority

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type RestoredLicenseLookup struct {
	OperationID          uuid.UUID
	EnvironmentID        uuid.UUID
	RestoreReceiptID     uuid.UUID
	RestoreReceiptSHA256 string
}

type RestoredLicenseState struct {
	Found                  bool
	OperationID            uuid.UUID
	RestoreReceiptID       uuid.UUID
	RestoreReceiptSHA256   string
	LicenseBindingID       uuid.UUID
	LicenseID              uuid.UUID
	LicenseeOrganizationID uuid.UUID
	EnvironmentID          uuid.UUID
	PreviousSequence       int64
	PreviousSHA256         string
	AcceptedSequence       int64
	AcceptedSHA256         string
	AuthorityRevision      int64
	DatabaseNow            time.Time
}

type RestoredLicenseReceiptStore interface {
	GetRestoredLicenseReceipt(
		context.Context,
		uuid.UUID,
		uuid.UUID,
	) (RestoredLicenseState, error)
}

func GetRestoredLicenseReceipt(
	ctx context.Context,
	store RestoredLicenseReceiptStore,
	lookup RestoredLicenseLookup,
) (RestoredLicenseState, error) {
	if ctx == nil || ctx.Err() != nil || store == nil ||
		lookup.OperationID == uuid.Nil ||
		lookup.OperationID.Version() != uuid.Version(7) ||
		lookup.EnvironmentID == uuid.Nil || lookup.RestoreReceiptID == uuid.Nil ||
		!digestPattern.MatchString(lookup.RestoreReceiptSHA256) {
		return RestoredLicenseState{}, ErrAuthorityState
	}
	state, err := store.GetRestoredLicenseReceipt(
		ctx, lookup.OperationID, lookup.EnvironmentID,
	)
	if err != nil {
		return RestoredLicenseState{}, err
	}
	if !state.Found {
		if !validRestoredLicenseClock(state.DatabaseNow) {
			return RestoredLicenseState{}, ErrAuthorityState
		}
		return state, nil
	}
	if state.OperationID != lookup.OperationID ||
		state.EnvironmentID != lookup.EnvironmentID ||
		state.RestoreReceiptID != lookup.RestoreReceiptID ||
		state.RestoreReceiptSHA256 != lookup.RestoreReceiptSHA256 ||
		state.LicenseBindingID == uuid.Nil || state.LicenseID == uuid.Nil ||
		state.LicenseeOrganizationID == uuid.Nil ||
		state.PreviousSequence < 1 ||
		state.AcceptedSequence <= state.PreviousSequence ||
		!digestPattern.MatchString(state.PreviousSHA256) ||
		!digestPattern.MatchString(state.AcceptedSHA256) ||
		state.AuthorityRevision < 1 ||
		!validRestoredLicenseClock(state.DatabaseNow) {
		return RestoredLicenseState{}, ErrWrongBinding
	}
	return state, nil
}

func validRestoredLicenseClock(value time.Time) bool {
	return !value.IsZero() && value.Location() == time.UTC
}
