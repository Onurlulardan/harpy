package devjournal

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"

	canonicaljsonv1 "github.com/Onurlulardan/harpy/contracts/canonicaljson/v1"
	"github.com/Onurlulardan/harpy/core/internal/modules/platformauthority/domain/artifact"
)

func decodeBindingProof(document []byte) (artifact.BindingProof, error) {
	if err := canonicaljsonv1.Validate(document); err != nil {
		return artifact.BindingProof{}, fmt.Errorf("binding proof JSON is invalid")
	}
	decoder := json.NewDecoder(bytes.NewReader(document))
	decoder.DisallowUnknownFields()
	var value artifact.BindingProof
	if decoder.Decode(&value) != nil || decoder.Decode(&struct{}{}) != io.EOF {
		return artifact.BindingProof{}, fmt.Errorf("binding proof schema is invalid")
	}
	return value, nil
}
