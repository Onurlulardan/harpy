package postrestorereactivationv1

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"

	canonicaljsonv1 "github.com/Onurlulardan/harpy/contracts/canonicaljson/v1"
	"github.com/google/uuid"
)

const (
	AwaitSchema          = "operator-post-restore-reactivation-await.v1"
	MaximumDocumentBytes = 16 * 1024
)

type Await struct {
	Schema             string `json:"schema"`
	EnvironmentID      string `json:"environment_id"`
	RestoreOperationID string `json:"restore_operation_id"`
}

func BuildAwait(environmentID, restoreOperationID string) ([]byte, error) {
	value := Await{
		Schema: AwaitSchema, EnvironmentID: environmentID,
		RestoreOperationID: restoreOperationID,
	}
	if !value.valid() {
		return nil, fmt.Errorf("post-restore reactivation await input is invalid")
	}
	raw, err := json.Marshal(value)
	if err != nil {
		return nil, fmt.Errorf("post-restore reactivation await encoding failed")
	}
	return canonicaljsonv1.CanonicalizeObject(raw, MaximumDocumentBytes)
}

func ParseAwait(document []byte) (Await, error) {
	canonical, err := canonicaljsonv1.CanonicalizeObject(
		document, MaximumDocumentBytes,
	)
	if err != nil || !bytes.Equal(canonical, document) {
		return Await{}, fmt.Errorf("post-restore reactivation await input is invalid")
	}
	var value Await
	decoder := json.NewDecoder(bytes.NewReader(canonical))
	decoder.DisallowUnknownFields()
	if decoder.Decode(&value) != nil || decoder.Decode(new(any)) != io.EOF ||
		!value.valid() {
		return Await{}, fmt.Errorf("post-restore reactivation await input is invalid")
	}
	return value, nil
}

func (value Await) valid() bool {
	return value.Schema == AwaitSchema &&
		validUUID(value.EnvironmentID) && validUUID(value.RestoreOperationID)
}

func validUUID(value string) bool {
	parsed, err := uuid.Parse(value)
	return err == nil && parsed != uuid.Nil && parsed.String() == value
}
