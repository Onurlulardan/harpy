package catalogcomponentv2

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"regexp"
	"strconv"
	"time"

	"github.com/google/uuid"
)

const canonicalTimeLayout = "2006-01-02T15:04:05.000000Z07:00"

var packageSlugPattern = regexp.MustCompile(`^[a-z][a-z0-9-]{1,62}$`)
var idempotencyPattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._:-]{7,127}$`)

func (value PurgeAuthority) Valid() bool {
	requested := canonicalTime(value.RequestedAt)
	return validStableUUID(value.TenantID) && validStableUUID(value.PackageID) &&
		validUUID(value.OperationID) && validUUID(value.ActorID) &&
		validUUID(value.RequestedBy) && validUUID(value.ApprovedBy) &&
		validUUID(value.ApprovalReceiptID) && validDigest(value.ApprovalSHA256) &&
		validUUID(value.BackupReceiptID) && value.ActorID == value.RequestedBy &&
		value.RequestedBy != value.ApprovedBy && packageSlugPattern.MatchString(value.PackageSlug) &&
		value.ConfirmationSlug == value.PackageSlug &&
		value.BackupState == BackupReady && validDigest(value.BackupSHA256) &&
		value.Reason != "" && len(value.Reason) <= 1024 && value.ReferenceEpoch > 0 &&
		value.AuthorizationTokenVersion > 0 && value.AuthorizationVersion > 0 &&
		value.AuthorizationClock > 0 &&
		validStableUUID(value.AuthorizationPermissionID) &&
		validDigest(value.AuthorizationEvidenceSHA256) &&
		validDigest(value.StepUpPurposeSHA256) &&
		idempotencyPattern.MatchString(value.IdempotencyKey) &&
		canonicalTime(value.StepUpAuthenticatedAt) && requested &&
		!value.StepUpAuthenticatedAt.After(value.RequestedAt) &&
		value.RequestedAt.Sub(value.StepUpAuthenticatedAt) <= 10*time.Minute &&
		canonicalTime(value.ApprovalStepUpAt) && canonicalTime(value.ApprovedAt) &&
		!value.ApprovalStepUpAt.After(value.ApprovedAt) &&
		value.ApprovedAt.Sub(value.ApprovalStepUpAt) <= 10*time.Minute &&
		canonicalTime(value.ApprovalExpiresAt) &&
		!value.ApprovalExpiresAt.Before(value.ApprovedAt) &&
		!value.RequestedAt.Before(value.ApprovedAt) &&
		!value.RequestedAt.After(value.ApprovalExpiresAt) &&
		canonicalTime(value.CoolingUntil) &&
		!value.CoolingUntil.Before(value.RequestedAt.Add(24*time.Hour)) &&
		validDigest(value.AuthoritySHA256) &&
		value.AuthoritySHA256 == AuthoritySHA256(value)
}

func AuthoritySHA256(value PurgeAuthority) string {
	return digestFields(
		"catalog.component-purge-authority.v2", value.TenantID, value.PackageID,
		value.PackageSlug, value.OperationID, value.ActorID, value.RequestedBy,
		value.ApprovedBy, value.ApprovalReceiptID, value.ApprovalSHA256,
		value.BackupReceiptID, value.BackupState,
		value.BackupSHA256, value.ConfirmationSlug, value.Reason,
		value.IdempotencyKey,
		strconv.FormatUint(value.ReferenceEpoch, 10),
		strconv.FormatInt(value.AuthorizationTokenVersion, 10),
		strconv.FormatInt(value.AuthorizationVersion, 10),
		strconv.FormatInt(value.AuthorizationClock, 10),
		value.AuthorizationPermissionID, value.AuthorizationEvidenceSHA256,
		value.StepUpPurposeSHA256,
		formatTime(value.StepUpAuthenticatedAt), formatTime(value.ApprovalStepUpAt),
		formatTime(value.ApprovedAt), formatTime(value.ApprovalExpiresAt),
		formatTime(value.RequestedAt),
		formatTime(value.CoolingUntil),
	)
}

func (command LifecycleCommand) Valid() bool {
	return validStableUUID(command.TenantID) && validStableUUID(command.PackageID) &&
		packageSlugPattern.MatchString(command.PackageSlug) &&
		validUUID(command.OperationID) && validUUID(command.ActorID) &&
		validComponent(command.Component) &&
		(command.Action == "retire" || command.Action == "restore") &&
		command.Reason != "" && len(command.Reason) <= 1024 &&
		command.ReferenceEpoch > 0 && idempotencyPattern.MatchString(command.IdempotencyKey)
}

func validComponent(value Component) bool {
	return validUUID(value.ID) && value.Owner != "" && value.Generation > 0 &&
		value.Contract != "" && validDigest(value.ContentSHA256)
}

func validUUID(value string) bool {
	parsed, err := uuid.Parse(value)
	return err == nil && parsed != uuid.Nil && parsed.Version() == 7 &&
		parsed.Variant() == uuid.RFC4122 && parsed.String() == value
}

func validStableUUID(value string) bool {
	parsed, err := uuid.Parse(value)
	return err == nil && parsed != uuid.Nil && parsed.String() == value
}

func validDigest(value string) bool {
	decoded, err := hex.DecodeString(value)
	return err == nil && len(decoded) == sha256.Size && value == hex.EncodeToString(decoded)
}

func canonicalTime(value time.Time) bool {
	return !value.IsZero() && value.Location() == time.UTC && value.Nanosecond()%1000 == 0
}

func formatTime(value time.Time) string {
	if value.IsZero() {
		return ""
	}
	return value.UTC().Format(canonicalTimeLayout)
}

func digestFields(values ...string) string {
	hash := sha256.New()
	var size [4]byte
	for _, value := range values {
		binary.BigEndian.PutUint32(size[:], uint32(len(value)))
		_, _ = hash.Write(size[:])
		_, _ = hash.Write([]byte(value))
	}
	return hex.EncodeToString(hash.Sum(nil))
}
