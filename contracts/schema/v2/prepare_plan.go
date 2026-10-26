package schemav2

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"

	canonicaljsonv1 "github.com/Onurlulardan/harpy/contracts/canonicaljson/v1"
)

type PreparedPlan struct {
	canonical []byte
	sha256    string
	steps     []PreparedStep
}

type PreparedStep struct {
	ordinal   int
	canonical []byte
	sha256    string
}

func PreparePlan(plan Plan) (PreparedPlan, error) {
	if plan.ContractVersion != PlanContractVersion || plan.Owner != "forms" ||
		plan.Compiler.Version != CompilerVersion ||
		plan.Compiler.PostgreSQLMajor != RequiredPostgreSQLMajor ||
		len(plan.Steps) < 1 || len(plan.Steps) > MaximumPlanSteps {
		return PreparedPlan{}, fmt.Errorf("schema plan contract is invalid")
	}
	for ordinal, step := range plan.Steps {
		if step.Ordinal != ordinal {
			return PreparedPlan{}, fmt.Errorf("schema plan ordinals are not gapless")
		}
	}
	canonical, err := marshalCanonical(plan)
	if err != nil || len(canonical) > MaximumSchemaPlanBytes {
		return PreparedPlan{}, fmt.Errorf("canonical schema plan is invalid")
	}
	planSHA256 := digest(canonical)
	steps := make([]PreparedStep, len(plan.Steps))
	for ordinal, step := range plan.Steps {
		prepared, prepareErr := PrepareStep(planSHA256, step)
		if prepareErr != nil {
			return PreparedPlan{}, fmt.Errorf("prepare step %d: %w", ordinal, prepareErr)
		}
		steps[ordinal] = prepared
	}
	return PreparedPlan{canonical: canonical, sha256: planSHA256, steps: steps}, nil
}

func PrepareStep(planSHA256 string, step Step) (PreparedStep, error) {
	decoded, err := hex.DecodeString(planSHA256)
	if err != nil || len(decoded) != sha256.Size || hex.EncodeToString(decoded) != planSHA256 ||
		step.Ordinal < 0 || step.Ordinal >= MaximumPlanSteps {
		return PreparedStep{}, fmt.Errorf("schema step identity is invalid")
	}
	canonical, err := marshalCanonical(struct {
		PlanSHA256 string `json:"plan_sha256"`
		Step       Step   `json:"step"`
	}{PlanSHA256: planSHA256, Step: step})
	if err != nil {
		return PreparedStep{}, err
	}
	return PreparedStep{
		ordinal: step.Ordinal, canonical: canonical, sha256: digest(canonical),
	}, nil
}

func marshalCanonical(value any) ([]byte, error) {
	encoded, err := json.Marshal(value)
	if err != nil {
		return nil, fmt.Errorf("encode schema contract JSON: %w", err)
	}
	canonical, err := canonicaljsonv1.CanonicalizeObject(encoded, MaximumSchemaPlanBytes)
	if err != nil {
		return nil, fmt.Errorf("canonicalize schema contract JSON: %w", err)
	}
	return canonical, nil
}

func digest(document []byte) string {
	value := sha256.Sum256(document)
	return hex.EncodeToString(value[:])
}

func (plan PreparedPlan) CanonicalJSON() []byte {
	return append([]byte(nil), plan.canonical...)
}

func (plan PreparedPlan) SHA256() string { return plan.sha256 }

func (plan PreparedPlan) Steps() []PreparedStep {
	return append([]PreparedStep(nil), plan.steps...)
}

func (step PreparedStep) Ordinal() int { return step.ordinal }

func (step PreparedStep) CanonicalPreimage() []byte {
	return append([]byte(nil), step.canonical...)
}

func (step PreparedStep) SHA256() string { return step.sha256 }
