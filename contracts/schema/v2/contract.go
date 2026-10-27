package schemav2

const (
	PlanContractVersion                 = "forms-schema-plan.v2"
	CompilerVersion                     = "2.0.0"
	RequiredPostgreSQLMajor             = 18
	MaximumSchemaPlanBytes              = 1 << 20
	MaximumPlanSteps                    = 2000
	MaximumBackfillBatch                = 1000
	MaximumRelationEdgesPerSource       = 100
	FormSystemColumnCount               = 8
	MaximumFieldGeneration        int64 = 2147483647
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
