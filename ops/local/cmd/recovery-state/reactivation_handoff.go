package main

import (
	"encoding/json"
	"fmt"
	"net/url"
	"path/filepath"
	"time"

	operatorv1 "github.com/Onurlulardan/harpy/contracts/operator/v1"
	"github.com/google/uuid"
)

const reactivationHandoffName = "post-restore-reactivation-handoff.v1.json"

type reactivationHandoff struct {
	Schema             string `json:"schema"`
	EnvironmentID      string `json:"environment_id"`
	RestoreOperationID string `json:"restore_operation_id"`
	URL                string `json:"url"`
	CreatedAt          string `json:"created_at"`
}

func buildReactivationHandoff(
	result operatorv1.IdentityReconcileResult,
	baseURL string,
	createdAt time.Time,
) (reactivationHandoff, error) {
	parsed, err := url.Parse(baseURL)
	if err != nil || parsed.Scheme != "http" && parsed.Scheme != "https" ||
		parsed.Host == "" || parsed.User != nil || parsed.RawQuery != "" ||
		parsed.Fragment != "" || parsed.Path != "" && parsed.Path != "/" ||
		createdAt.IsZero() || result.Validate() != nil {
		return reactivationHandoff{}, fmt.Errorf("reactivation handoff input is invalid")
	}
	parsed.Path = "/post-restore-reactivation"
	query := parsed.Query()
	query.Set("environment_id", result.EnvironmentID)
	query.Set("restore_operation_id", result.RestoreOperationID)
	parsed.RawQuery = query.Encode()
	return reactivationHandoff{
		Schema:        "post-restore-reactivation-handoff.v1",
		EnvironmentID: result.EnvironmentID, RestoreOperationID: result.RestoreOperationID,
		URL: parsed.String(), CreatedAt: createdAt.UTC().Format(time.RFC3339Nano),
	}, nil
}

func (handoff reactivationHandoff) valid() bool {
	parsed, err := url.Parse(handoff.URL)
	createdAt, timeErr := time.Parse(time.RFC3339Nano, handoff.CreatedAt)
	query := parsed.Query()
	return handoff.Schema == "post-restore-reactivation-handoff.v1" &&
		canonicalHandoffUUID(handoff.EnvironmentID) &&
		canonicalHandoffUUID(handoff.RestoreOperationID) && err == nil &&
		(parsed.Scheme == "http" || parsed.Scheme == "https") && parsed.Host != "" &&
		parsed.User == nil && parsed.Fragment == "" &&
		parsed.Path == "/post-restore-reactivation" &&
		query.Get("environment_id") == handoff.EnvironmentID &&
		query.Get("restore_operation_id") == handoff.RestoreOperationID &&
		len(query) == 2 && timeErr == nil && createdAt.Location() == time.UTC
}

func canonicalHandoffUUID(value string) bool {
	parsed, err := uuid.Parse(value)
	return err == nil && parsed != uuid.Nil && parsed.String() == value
}

func encodeReactivationHandoff(handoff reactivationHandoff) ([]byte, error) {
	if !handoff.valid() {
		return nil, fmt.Errorf("reactivation handoff is invalid")
	}
	return json.Marshal(handoff)
}

func reactivationHandoffPath(directory string) string {
	return filepath.Join(directory, reactivationHandoffName)
}
