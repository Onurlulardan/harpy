package lifecycle

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"regexp"
	"time"

	"github.com/google/uuid"
)

const (
	purgeApprovalAvailable = "AVAILABLE"
	purgeApprovalConsumed  = "CONSUMED"
)

var purgeSlugPattern = regexp.MustCompile(`^[a-z][a-z0-9-]{1,62}$`)

type purgeApprovalRequestDocument struct {
	TenantID, InstallationID, PackageID, OperationID, BackupReceiptID string
	RequestedBy, Reason, ConfirmationSlug, PurgeIdempotencyKey        string
}

func (value PurgeApprovalReceipt) Valid() bool {
	if !validV7(value.ID) || !validStableUUID(value.TenantID) ||
		!validV7(value.InstallationID) || !validStableUUID(value.PackageID) ||
		!validV7(value.OperationID) || !validV7(value.RequestedBy) ||
		!validV7(value.ApprovedBy) || value.RequestedBy == value.ApprovedBy ||
		!validV7(value.BackupReceiptID) || !validDigest(value.BackupSHA256) ||
		!purgeSlugPattern.MatchString(value.ConfirmationSlug) ||
		value.Reason == "" || len(value.Reason) > 1024 ||
		!validIdempotency(value.PurgeIdempotencyKey) ||
		!validDigest(value.RequestSHA256) || !value.validRequestSHA256() ||
		!validIdempotency(value.ApprovalIdempotencyKey) ||
		!canonicalApprovalTime(value.StepUpAuthenticatedAt) ||
		!canonicalApprovalTime(value.ApprovedAt) ||
		!canonicalApprovalTime(value.ExpiresAt) ||
		value.StepUpAuthenticatedAt.After(value.ApprovedAt) ||
		value.ApprovedAt.Sub(value.StepUpAuthenticatedAt) > 10*time.Minute ||
		!value.ExpiresAt.After(value.ApprovedAt) ||
		!validDigest(value.ApprovalSHA256) ||
		value.ApprovalSHA256 != approvalReceiptSHA256(value) {
		return false
	}
	if value.State == purgeApprovalAvailable {
		return value.ConsumedOperationID == "" &&
			value.ConsumedAuthoritySHA256 == "" && value.ConsumedAt.IsZero()
	}
	return value.State == purgeApprovalConsumed &&
		validV7(value.ConsumedOperationID) &&
		validDigest(value.ConsumedAuthoritySHA256) &&
		canonicalApprovalTime(value.ConsumedAt)
}

func (value PurgeApprovalReceipt) validRequestSHA256() bool {
	digest, err := approvalRequestDocumentSHA(purgeApprovalRequestDocument{
		TenantID: value.TenantID, InstallationID: value.InstallationID,
		PackageID: value.PackageID, OperationID: value.OperationID,
		BackupReceiptID: value.BackupReceiptID, RequestedBy: value.RequestedBy,
		Reason: value.Reason, ConfirmationSlug: value.ConfirmationSlug,
		PurgeIdempotencyKey: value.PurgeIdempotencyKey,
	})
	return err == nil && digest == value.RequestSHA256
}

func approvalRequestDocumentSHA(value purgeApprovalRequestDocument) (string, error) {
	return requestDigest(value)
}

func approvalReceiptSHA256(value PurgeApprovalReceipt) string {
	return digestApprovalFields(
		"catalog.purge-approval-receipt.v2", value.ID, value.TenantID,
		value.InstallationID, value.PackageID, value.OperationID,
		value.RequestedBy, value.ApprovedBy, value.BackupReceiptID,
		value.BackupSHA256, value.ConfirmationSlug, value.Reason,
		value.PurgeIdempotencyKey, value.RequestSHA256,
		value.ApprovalIdempotencyKey,
		formatApprovalTime(value.StepUpAuthenticatedAt),
		formatApprovalTime(value.ApprovedAt), formatApprovalTime(value.ExpiresAt),
	)
}

func digestApprovalFields(values ...string) string {
	hash := sha256.New()
	var size [4]byte
	for _, value := range values {
		binary.BigEndian.PutUint32(size[:], uint32(len(value)))
		_, _ = hash.Write(size[:])
		_, _ = hash.Write([]byte(value))
	}
	return hex.EncodeToString(hash.Sum(nil))
}

func validV7(value string) bool {
	parsed, err := uuid.Parse(value)
	return err == nil && validUUIDv7(parsed) && parsed.String() == value
}

func validStableUUID(value string) bool {
	_, err := uuidFromCanonical(value)
	return err == nil
}

func validUUIDv7(value uuid.UUID) bool {
	return value != uuid.Nil && value.Version() == uuid.Version(7) &&
		value.Variant() == uuid.RFC4122
}

func canonicalApprovalTime(value time.Time) bool {
	return !value.IsZero() && value.Location() == time.UTC && value.Nanosecond()%1000 == 0
}

func formatApprovalTime(value time.Time) string {
	return value.UTC().Format("2006-01-02T15:04:05.000000Z07:00")
}
