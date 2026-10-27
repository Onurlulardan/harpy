package application

import (
	"bytes"
	"fmt"

	schemav1 "github.com/Onurlulardan/harpy/contracts/schema/v1"
	"github.com/Onurlulardan/harpy/core/internal/platform/schemaexecutor"
	submissiondomain "github.com/Onurlulardan/harpy/core/internal/platform/schemasubmission/domain"
	"github.com/google/uuid"
)

func prepare(command SubmitCommand) (PreparedSubmission, error) {
	if _, err := submissiondomain.NewJobID(command.ProposedJobID); err != nil {
		return PreparedSubmission{}, ErrInvalid
	}
	if command.PlanV2 != nil {
		if command.Plan.ContractVersion != "" || command.Plan.TenantID != "" ||
			command.Plan.Owner != "" || command.Plan.Steps != nil {
			return PreparedSubmission{}, ErrInvalid
		}
		return prepareV2(command.ProposedJobID, *command.PlanV2)
	}
	tenantID, err := parseUUIDv7(command.Plan.TenantID)
	if err != nil {
		return PreparedSubmission{}, ErrInvalid
	}
	resourceID, err := parseUUIDv7(command.Plan.Resource.ID)
	if err != nil {
		return PreparedSubmission{}, ErrInvalid
	}
	validated, err := schemaexecutor.BuildPlan(command.Plan)
	if err != nil {
		return PreparedSubmission{}, fmt.Errorf("%w: %v", ErrInvalid, err)
	}
	prepared, err := schemav1.PreparePlan(command.Plan)
	if err != nil || !bytes.Equal(validated.CanonicalJSON(), prepared.CanonicalJSON()) ||
		validated.SHA256() != prepared.SHA256() {
		return PreparedSubmission{}, ErrInvalid
	}
	stepDocuments, err := prepareStepDocuments(prepared, validated)
	if err != nil {
		return PreparedSubmission{}, fmt.Errorf("%w: %v", ErrInvalid, err)
	}
	return PreparedSubmission{
		jobID: command.ProposedJobID, tenantID: tenantID,
		resourceKind: string(command.Plan.Resource.Kind), resourceID: resourceID,
		resourceRevision: command.Plan.Resource.Revision,
		contractVersion:  command.Plan.ContractVersion,
		compilerVersion:  command.Plan.Compiler.Version,
		postgresqlMajor:  command.Plan.Compiler.PostgreSQLMajor,
		canonicalPlan:    prepared.CanonicalJSON(), planSHA256: prepared.SHA256(),
		canonicalSteps: stepDocuments,
	}, nil
}

func parseUUIDv7(value string) (uuid.UUID, error) {
	parsed, err := uuid.Parse(value)
	if err != nil || parsed.String() != value || !submissiondomain.ValidUUIDv7(parsed) {
		return uuid.Nil, ErrInvalid
	}
	return parsed, nil
}

func prepareStepDocuments(
	prepared schemav1.PreparedPlan,
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
