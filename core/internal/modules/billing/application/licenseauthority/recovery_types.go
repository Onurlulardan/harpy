package licenseauthority

import (
	"context"

	"github.com/google/uuid"
)

type RestoredLicenseCommit struct {
	OperationID            uuid.UUID
	EnvironmentID          uuid.UUID
	RestoreReceiptID       uuid.UUID
	RestoreReceiptSHA256   string
	LicenseBindingID       uuid.UUID
	LicenseID              uuid.UUID
	LicenseeOrganizationID uuid.UUID
	PreviousSequence       int64
	PreviousSHA256         string
	NextSequence           int64
	NextSHA256             string
	ExpectedRevision       int64
	Snapshot               VerifiedSnapshot
}

type RestoredLicenseReceipt struct {
	OperationID       uuid.UUID
	RestoreReceiptID  uuid.UUID
	LicenseBindingID  uuid.UUID
	PreviousSequence  int64
	PreviousSHA256    string
	AcceptedSequence  int64
	AcceptedSHA256    string
	AuthorityRevision int64
	Replay            bool
}

type RestoredLicenseStore interface {
	CommitRestoredLicense(context.Context, RestoredLicenseCommit) (RestoredLicenseReceipt, error)
}
