package schemav1

const (
	PlanContractVersion          = "forms-schema-plan.v1"
	InitialCompilerVersion       = "1.0.0"
	RequiredPostgreSQLMajor      = 18
	MaximumSchemaPlanBytes       = 1 << 20
	MaximumPlanSteps             = 2000
	MaximumBackfillBatch         = 1000
	InitialFormSystemColumnCount = 5
)

type ResourceKind string

const (
	ResourceFormStorage ResourceKind = "form_storage"
	ResourceFormPurge   ResourceKind = "form_purge"
)

type Compiler struct {
	Version         string `json:"version"`
	PostgreSQLMajor int    `json:"postgresql_major"`
}

type Resource struct {
	Kind     ResourceKind `json:"kind"`
	ID       string       `json:"id"`
	Revision int64        `json:"revision"`
}
