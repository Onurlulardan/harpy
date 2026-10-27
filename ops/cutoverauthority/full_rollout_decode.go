package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"

	corecutover "github.com/Onurlulardan/harpy/core/public/cutover"
	"github.com/google/uuid"
)

const fullRolloutEvidenceContract = "full-rollout-completion-evidence.v1"

func loadFullRolloutCompletion(path string) (
	corecutover.FullRolloutCompletion, error,
) {
	raw, err := readArtifact(path, maximumArtifactBytes)
	if err != nil {
		return corecutover.FullRolloutCompletion{}, err
	}
	var document fullRolloutDocument
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if decoder.Decode(&document) != nil || decoder.Decode(&struct{}{}) != io.EOF {
		return corecutover.FullRolloutCompletion{}, fmt.Errorf("completion evidence is invalid")
	}
	canonical, err := json.Marshal(document)
	if err != nil || !bytes.Equal(raw, canonical) ||
		document.ContractVersion != fullRolloutEvidenceContract {
		return corecutover.FullRolloutCompletion{}, fmt.Errorf("completion evidence is noncanonical")
	}
	receiptID, err := parseFullRolloutUUID(document.ReceiptID)
	if err != nil {
		return corecutover.FullRolloutCompletion{}, err
	}
	environmentID, err := parseFullRolloutUUID(document.EnvironmentID)
	if err != nil {
		return corecutover.FullRolloutCompletion{}, err
	}
	recordedAt, err := parseFullRolloutTime(document.RecordedAt)
	if err != nil {
		return corecutover.FullRolloutCompletion{}, err
	}
	evidence, err := document.evidence()
	if err != nil {
		return corecutover.FullRolloutCompletion{}, err
	}
	return corecutover.FullRolloutCompletion{ReceiptID: receiptID,
		EnvironmentID: environmentID, RecordedAt: recordedAt, Evidence: evidence}, nil
}

func parseFullRolloutUUID(value string) (uuid.UUID, error) {
	parsed, err := uuid.Parse(value)
	if err != nil || parsed == uuid.Nil || parsed.String() != value {
		return uuid.Nil, fmt.Errorf("completion evidence UUID is invalid")
	}
	return parsed, nil
}
