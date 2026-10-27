package schemaexecutor

import (
	"bytes"
	"fmt"

	schemav1 "github.com/Onurlulardan/harpy/contracts/schema/v1"
	"github.com/Onurlulardan/harpy/internal/attestation"
)

func validateStepRow(planHash string, step schemav1.Step, row StepRecord) (string, error) {
	if row.OperationID != string(step.OperationID) ||
		row.TransactionClass != string(step.TransactionClass) ||
		row.ExecutorKind != string(step.ExecutorKind) {
		return "", fmt.Errorf("persisted execution identity differs")
	}
	if err := compareFragment(row.IdentifierMapping, step.IdentifierMapping); err != nil {
		return "", fmt.Errorf("identifier mapping differs: %w", err)
	}
	if err := compareFragment(row.Arguments, step.Arguments); err != nil {
		return "", fmt.Errorf("arguments differ: %w", err)
	}
	canonical, err := attestation.MarshalCanonicalSchemaPlan(struct {
		PlanSHA256 string        `json:"plan_sha256"`
		Step       schemav1.Step `json:"step"`
	}{PlanSHA256: planHash, Step: step})
	if err != nil {
		return "", err
	}
	hash, err := attestation.DigestSchemaPlan(canonical)
	if err != nil {
		return "", err
	}
	if row.StepSHA256 != hash {
		return "", fmt.Errorf("step hash differs")
	}
	return hash, nil
}

func compareFragment(document []byte, value any) error {
	canonicalDocument, err := canonicalizePlan(document)
	if err != nil {
		return err
	}
	canonicalValue, err := attestation.MarshalCanonicalSchemaPlan(value)
	if err != nil {
		return err
	}
	if !bytes.Equal(canonicalDocument, canonicalValue) {
		return fmt.Errorf("canonical value mismatch")
	}
	return nil
}
