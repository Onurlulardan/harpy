package application

import (
	"encoding/hex"

	"github.com/google/uuid"
)

type PreparedSubmission struct {
	jobID            uuid.UUID
	tenantID         uuid.UUID
	resourceKind     string
	resourceID       uuid.UUID
	resourceRevision int64
	contractVersion  string
	compilerVersion  string
	postgresqlMajor  int
	canonicalPlan    []byte
	planSHA256       string
	canonicalSteps   [][]byte
}

func (value PreparedSubmission) JobID() uuid.UUID        { return value.jobID }
func (value PreparedSubmission) TenantID() uuid.UUID     { return value.tenantID }
func (value PreparedSubmission) ResourceKind() string    { return value.resourceKind }
func (value PreparedSubmission) ResourceID() uuid.UUID   { return value.resourceID }
func (value PreparedSubmission) ResourceRevision() int64 { return value.resourceRevision }
func (value PreparedSubmission) ContractVersion() string { return value.contractVersion }
func (value PreparedSubmission) CompilerVersion() string { return value.compilerVersion }
func (value PreparedSubmission) PostgreSQLMajor() int    { return value.postgresqlMajor }
func (value PreparedSubmission) PlanSHA256() string      { return value.planSHA256 }

func (value PreparedSubmission) CanonicalPlan() []byte {
	return append([]byte(nil), value.canonicalPlan...)
}

func (value PreparedSubmission) CanonicalSteps() [][]byte {
	steps := make([][]byte, len(value.canonicalSteps))
	for index := range value.canonicalSteps {
		steps[index] = append([]byte(nil), value.canonicalSteps[index]...)
	}
	return steps
}

func (value PreparedSubmission) CanonicalStepHex() []string {
	steps := make([]string, len(value.canonicalSteps))
	for index := range value.canonicalSteps {
		steps[index] = hex.EncodeToString(value.canonicalSteps[index])
	}
	return steps
}
