package schemaexecutor

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"unicode/utf8"

	schemav1 "github.com/Onurlulardan/harpy/contracts/schema/v1"
	"github.com/Onurlulardan/harpy/internal/attestation"
)

func canonicalizePlan(document []byte) ([]byte, error) {
	if len(document) < 2 || len(document) > schemav1.MaximumSchemaPlanBytes ||
		!utf8.Valid(document) {
		return nil, fmt.Errorf("plan document size or encoding is invalid")
	}
	if err := rejectDuplicateMembers(document); err != nil {
		return nil, err
	}
	canonical, err := attestation.CanonicalizeSchemaPlan(document)
	if err != nil {
		return nil, fmt.Errorf("canonicalize plan: %w", err)
	}
	if len(canonical) > schemav1.MaximumSchemaPlanBytes {
		return nil, fmt.Errorf("canonical plan exceeds size limit")
	}
	return canonical, nil
}

func decodeStrict(document []byte, target any) error {
	decoder := json.NewDecoder(bytes.NewReader(document))
	decoder.DisallowUnknownFields()
	decoder.UseNumber()
	if err := decoder.Decode(target); err != nil {
		return err
	}
	var trailing any
	err := decoder.Decode(&trailing)
	if err == nil {
		return fmt.Errorf("JSON has trailing value")
	}
	if err != io.EOF {
		return fmt.Errorf("JSON trailing data is invalid: %w", err)
	}
	return nil
}
