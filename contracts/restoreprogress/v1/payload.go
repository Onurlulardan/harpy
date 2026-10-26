package restoreprogressv1

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"

	canonicaljsonv1 "github.com/Onurlulardan/harpy/contracts/canonicaljson/v1"
)

func Build(input Input) ([]byte, error) {
	payload := Payload{
		Schema: Schema, EnvironmentID: input.EnvironmentID,
		OperationID: input.OperationID, AttemptID: input.AttemptID,
		PlanSHA256: input.PlanSHA256, State: input.State,
		PreviousDatabaseSystemID:      input.PreviousDatabaseSystemID,
		NewDatabaseSystemID:           input.NewDatabaseSystemID,
		PreviousDatabaseLocatorSHA256: input.PreviousDatabaseLocatorSHA256,
		NewDatabaseLocatorSHA256:      input.NewDatabaseLocatorSHA256,
		DrainAckSHA256:                input.DrainAckSHA256,
		SwitchAckSHA256:               input.SwitchAckSHA256,
		FailureCode:                   input.FailureCode,
		OccurredAt:                    input.OccurredAt.Format(timeFormat),
	}
	if !payload.valid() {
		return nil, fmt.Errorf("restore progress input is invalid")
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("restore progress encoding failed")
	}
	return canonicaljsonv1.CanonicalizeObject(raw, MaximumDocumentBytes)
}

func Parse(document []byte) (Payload, error) {
	canonical, err := canonicaljsonv1.CanonicalizeObject(
		document, MaximumDocumentBytes,
	)
	if err != nil || !bytes.Equal(canonical, document) {
		return Payload{}, fmt.Errorf("restore progress is invalid")
	}
	var payload Payload
	decoder := json.NewDecoder(bytes.NewReader(canonical))
	decoder.DisallowUnknownFields()
	if decoder.Decode(&payload) != nil || decoder.Decode(new(any)) != io.EOF ||
		!payload.valid() {
		return Payload{}, fmt.Errorf("restore progress is invalid")
	}
	return payload, nil
}
