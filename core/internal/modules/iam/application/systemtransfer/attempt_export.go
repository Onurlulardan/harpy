package systemtransfer

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"

	canonicaljsonv1 "github.com/Onurlulardan/harpy/contracts/canonicaljson/v1"
	publicresource "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/systemresource"
	"github.com/google/uuid"
)

const systemExportMaximumQueryBytes = 262144

type ExportOwnerState string

const (
	ExportWaiting        ExportOwnerState = "waiting"
	ExportRunning        ExportOwnerState = "running"
	ExportReady          ExportOwnerState = "ready"
	ExportFailed         ExportOwnerState = "failed"
	ExportCancelled      ExportOwnerState = "cancelled"
	ExportExpired        ExportOwnerState = "expired"
	ExportOutcomeUnknown ExportOwnerState = "outcome_unknown"
)

type ExportAttempt struct {
	TenantID, ActorID, JobID, ExportID uuid.UUID
	AttemptID, LeaseOwnerID            uuid.UUID
	CatalogVersionID                   uuid.UUID
	Resource                           publicresource.Resource
	Format                             string
	State                              ExportOwnerState
	TokenVersion, AuthorizationVersion int64
	CatalogSHA256, QuerySHA256         [32]byte
	QueryDocument                      []byte
	AttemptGeneration                  int64
}

func (attempt ExportAttempt) Validate() error {
	if attempt.CheckOwner() != nil || !exportV7(attempt.AttemptID) ||
		!exportV7(attempt.LeaseOwnerID) || attempt.AttemptGeneration < 1 {
		return ErrExportAttemptInvalid
	}
	return nil
}

func (attempt ExportAttempt) CheckOwner() error {
	if attempt.TenantID == uuid.Nil || attempt.ActorID == uuid.Nil ||
		!exportV7(attempt.JobID) || !exportV7(attempt.ExportID) ||
		!exportV7(attempt.CatalogVersionID) || !attempt.Resource.Valid() ||
		!validExportFormat(attempt.Format) || !attempt.State.Valid() ||
		attempt.TokenVersion < 1 || attempt.AuthorizationVersion < 1 ||
		attempt.CatalogSHA256 == [32]byte{} || attempt.QuerySHA256 == [32]byte{} ||
		attempt.validateQuery() != nil {
		return ErrExportAttemptInvalid
	}
	return nil
}

func (state ExportOwnerState) Valid() bool {
	return state == ExportWaiting || state == ExportRunning || state == ExportReady ||
		state == ExportFailed || state == ExportCancelled || state == ExportExpired ||
		state == ExportOutcomeUnknown
}

func (attempt ExportAttempt) validateQuery() error {
	canonical, err := canonicaljsonv1.CanonicalizeObject(
		attempt.QueryDocument, systemExportMaximumQueryBytes,
	)
	if err != nil || !bytes.Equal(canonical, attempt.QueryDocument) ||
		sha256.Sum256(canonical) != attempt.QuerySHA256 {
		return ErrExportQueryInvalid
	}
	var object map[string]json.RawMessage
	if json.Unmarshal(canonical, &object) != nil {
		return ErrExportQueryInvalid
	}
	for key := range object {
		if key != "select" && key != "where" && key != "sort" && key != "search" {
			return ErrExportQueryInvalid
		}
	}
	var sorts []struct {
		Field string `json:"field"`
	}
	if json.Unmarshal(object["sort"], &sorts) != nil || len(sorts) < 1 || len(sorts) > 5 {
		return ErrExportQueryInvalid
	}
	for index, sort := range sorts {
		if sort.Field == "id" && index != len(sorts)-1 {
			return ErrExportQueryInvalid
		}
	}
	if sorts[len(sorts)-1].Field != "id" {
		return ErrExportQueryInvalid
	}
	return nil
}

func validExportFormat(value string) bool {
	return value == "csv" || value == "jsonl" || value == "xlsx"
}
