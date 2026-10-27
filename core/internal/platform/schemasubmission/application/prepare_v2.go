package application

import (
	"bytes"
	"fmt"

	schemav2 "github.com/Onurlulardan/harpy/contracts/schema/v2"
	"github.com/Onurlulardan/harpy/core/internal/platform/schemaexecutor"
	"github.com/google/uuid"
)

func prepareV2(jobID uuid.UUID, plan schemav2.Plan) (PreparedSubmission, error) {
	tenantID, err := parseUUIDv7(plan.TenantID)
	if err != nil {
		return PreparedSubmission{}, ErrInvalid
	}
	resourceID, err := parseUUIDv7(plan.Resource.ID)
	if err != nil {
		return PreparedSubmission{}, ErrInvalid
	}
	validated, err := schemaexecutor.BuildPlanV2(plan)
	if err != nil {
		return PreparedSubmission{}, fmt.Errorf("%w: %v", ErrInvalid, err)
	}
	prepared, err := schemav2.PreparePlan(plan)
	if err != nil || !bytes.Equal(validated.CanonicalJSON(), prepared.CanonicalJSON()) ||
		validated.SHA256() != prepared.SHA256() {
		return PreparedSubmission{}, ErrInvalid
	}
	stepDocuments, err := prepareStepDocumentsV2(prepared, validated)
	if err != nil {
		return PreparedSubmission{}, fmt.Errorf("%w: %v", ErrInvalid, err)
	}
	return PreparedSubmission{
		jobID: jobID, tenantID: tenantID,
		resourceKind: string(plan.Resource.Kind), resourceID: resourceID,
		resourceRevision: plan.Resource.Revision,
		contractVersion:  plan.ContractVersion, compilerVersion: plan.Compiler.Version,
		postgresqlMajor: plan.Compiler.PostgreSQLMajor,
		canonicalPlan:   prepared.CanonicalJSON(), planSHA256: prepared.SHA256(),
		canonicalSteps: stepDocuments,
	}, nil
}

func prepareStepDocumentsV2(
	prepared schemav2.PreparedPlan,
	validated schemaexecutor.PreparedPlan,
) ([][]byte, error) {
	steps, validatedSteps := prepared.Steps(), validated.Steps()
	if len(steps) != len(validatedSteps) {
		return nil, ErrInvalid
	}
	documents := make([][]byte, len(steps))
	for index, step := range steps {
		if step.Ordinal() != index || step.SHA256() != validatedSteps[index].SHA256() {
			return nil, ErrInvalid
		}
		documents[index] = step.CanonicalPreimage()
	}
	return documents, nil
}
