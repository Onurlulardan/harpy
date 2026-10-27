package async

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

const (
	AnalyticsDashboardExportDeleteJobKind = "analytics.dashboard.export.delete"
	AnalyticsDashboardExportExpireJobKind = "analytics.dashboard.export.expire"
)

var analyticsDashboardExportReceiptNamespace = uuid.MustParse(
	"8e7a6d02-df6d-4327-b179-160296b50b8e",
)

type AnalyticsDashboardExportMaintenanceUnknownCandidate struct {
	Kind                     string
	TenantID                 uuid.UUID
	TenantSchema             string
	ExportID                 uuid.UUID
	AuthorityGroupID         string
	AuthorityGroupGeneration int64
	CutoverGeneration        int64
	JobID                    uuid.UUID
	AttemptID                uuid.UUID
	AttemptNumber            int64
	LeaseGeneration          int64
	OriginalLeaseOwnerID     uuid.UUID
	EffectID                 uuid.UUID
	EffectRequestSHA256      [32]byte
	UnknownAt                time.Time
}

func (candidate AnalyticsDashboardExportMaintenanceUnknownCandidate) Valid() bool {
	maximum, validKind := analyticsDashboardExportMaintenanceMaximum(candidate.Kind)
	return validKind && candidate.TenantID != uuid.Nil &&
		validAnalyticsExportSchema(candidate.TenantSchema) &&
		validAnalyticsExportSchema(candidate.AuthorityGroupID) &&
		candidate.AuthorityGroupGeneration >= 1 && candidate.CutoverGeneration >= 1 &&
		validAnalyticsExportV7(candidate.ExportID) && validAnalyticsExportV7(candidate.JobID) &&
		validAnalyticsExportV7(candidate.AttemptID) && candidate.AttemptNumber >= 1 &&
		candidate.AttemptNumber <= maximum &&
		candidate.LeaseGeneration == candidate.AttemptNumber &&
		validAnalyticsExportV7(candidate.OriginalLeaseOwnerID) &&
		validAnalyticsExportV5(candidate.EffectID) &&
		candidate.EffectRequestSHA256 != [32]byte{} && !candidate.UnknownAt.IsZero()
}

type AnalyticsDashboardExportMaintenanceUnknownSource interface {
	NextAnalyticsDashboardExportMaintenanceUnknown(
		context.Context,
	) (AnalyticsDashboardExportMaintenanceUnknownCandidate, bool, error)
}

type AnalyticsDashboardExportSource interface {
	AnalyticsDashboardExportWorkerSource
	AnalyticsDashboardExportMaintenanceUnknownSource
}

func AnalyticsDashboardExportMaintenanceReceiptID(
	candidate AnalyticsDashboardExportMaintenanceUnknownCandidate,
) (uuid.UUID, error) {
	_, code, _, valid := analyticsDashboardExportMaintenanceOutcome(candidate, "deleted")
	if candidate.Kind == AnalyticsDashboardExportExpireJobKind {
		_, code, _, valid = analyticsDashboardExportMaintenanceOutcome(candidate, "expired")
	}
	if !valid {
		return uuid.Nil, ErrAnalyticsDashboardExportEvidenceInvalid
	}
	return uuid.NewSHA1(analyticsDashboardExportReceiptNamespace,
		[]byte(candidate.EffectID.String()+"\x00"+code)), nil
}

func AnalyticsDashboardExportMaintenanceReceiptSHA256(
	candidate AnalyticsDashboardExportMaintenanceUnknownCandidate,
	effectReceiptID uuid.UUID, terminalState string, resolvedAt time.Time,
) ([32]byte, error) {
	operation, code, resultReference, valid := analyticsDashboardExportMaintenanceOutcome(
		candidate, terminalState,
	)
	wantReceiptID, receiptErr := AnalyticsDashboardExportMaintenanceReceiptID(candidate)
	if !valid || receiptErr != nil || effectReceiptID != wantReceiptID ||
		!validAnalyticsExportV5(effectReceiptID) || resolvedAt.IsZero() ||
		!resolvedAt.After(candidate.UnknownAt) {
		return [32]byte{}, ErrAnalyticsDashboardExportEvidenceInvalid
	}
	resultDigest := sha256.Sum256([]byte(resultReference))
	fields := []string{
		"analytics-dashboard-export-maintenance-effect-receipt-v1",
		"effect-receipt.v1", effectReceiptID.String(), candidate.EffectID.String(),
		candidate.Kind, "1", candidate.TenantID.String(), "tenant",
		candidate.TenantID.String(), candidate.AuthorityGroupID,
		strconv.FormatInt(candidate.AuthorityGroupGeneration, 10), "analytics",
		strconv.FormatInt(candidate.CutoverGeneration, 10), "v2",
		"analytics.dashboard-export", operation,
		candidate.Kind + ":" + candidate.ExportID.String(), candidate.JobID.String(),
		candidate.JobID.String(), "succeeded", code, candidate.ExportID.String(),
		resultReference, hex.EncodeToString(resultDigest[:]),
		hex.EncodeToString(candidate.EffectRequestSHA256[:]),
		strconv.FormatInt(resolvedAt.UTC().UnixMicro(), 10),
	}
	return sha256.Sum256([]byte(strings.Join(fields, "\x00"))), nil
}

func analyticsDashboardExportMaintenanceOutcome(
	candidate AnalyticsDashboardExportMaintenanceUnknownCandidate, terminalState string,
) (string, string, string, bool) {
	if !candidate.Valid() {
		return "", "", "", false
	}
	operation, code := "delete", "ANALYTICS_EXPORT_DELETED"
	validState := terminalState == "deleted"
	if candidate.Kind == AnalyticsDashboardExportExpireJobKind {
		operation, code = "expire", "ANALYTICS_EXPORT_EXPIRY_COMPLETE"
		validState = terminalState == "expired" || terminalState == "cancelled" ||
			terminalState == "failed" || terminalState == "deleted"
	}
	resultReference := fmt.Sprintf("analytics-dashboard-export://%s/%s/%s",
		operation, candidate.ExportID, terminalState)
	return operation, code, resultReference, validState
}

func analyticsDashboardExportMaintenanceMaximum(kind string) (int64, bool) {
	switch kind {
	case AnalyticsDashboardExportDeleteJobKind:
		return 3, true
	case AnalyticsDashboardExportExpireJobKind:
		return 8, true
	default:
		return 0, false
	}
}

func validAnalyticsExportV5(value uuid.UUID) bool {
	return value != uuid.Nil && value.Version() == uuid.Version(5) &&
		value.Variant() == uuid.RFC4122
}
