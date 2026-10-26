package profilebackfill

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type mappingWire struct {
	Ordinal       int    `json:"ordinal"`
	SourceFieldID string `json:"source_field_id"`
	TargetFieldID string `json:"target_field_id"`
	ConverterKind string `json:"converter_kind"`
	SemanticSHA   string `json:"semantic_sha256"`
}

type pageWire struct {
	PageNumber        int     `json:"page_number"`
	JobID             string  `json:"job_id"`
	ScheduledRunAt    string  `json:"scheduled_run_at"`
	PayloadReference  string  `json:"payload_reference"`
	PayloadSHA        string  `json:"payload_sha256"`
	AfterResourceID   *string `json:"after_resource_id"`
	ThroughResourceID string  `json:"through_resource_id"`
	ExpectedCount     int     `json:"expected_count"`
	SourceSHA         string  `json:"source_sha256"`
	ExpectedTargetSHA string  `json:"expected_target_sha256"`
}

type planWire struct {
	ID                     string          `json:"id"`
	ResourceKind           string          `json:"resource_kind"`
	SourceSchemaID         string          `json:"source_schema_id"`
	TargetSchemaID         string          `json:"target_schema_id"`
	TargetVersion          int64           `json:"target_version"`
	TargetDefinition       json.RawMessage `json:"target_definition"`
	TargetDefinitionSHA256 string          `json:"target_definition_sha256"`
	ConverterVersion       string          `json:"converter_version"`
	SourceCount            int64           `json:"source_count"`
	ExpectedSourceSHA256   string          `json:"expected_source_sha256"`
	ExpectedTargetSHA256   string          `json:"expected_target_sha256"`
	PageSize               int             `json:"page_size"`
	InitialRollingSHA256   string          `json:"initial_rolling_sha256"`
	Mappings               []mappingWire   `json:"mappings"`
	Pages                  []pageWire      `json:"pages"`
}

func prepare(command CreateCommand, tenantID uuid.UUID) (PreparedPlan, error) {
	if command.Evidence.Validate() != nil || command.Plan.Validate() != nil || !stableID(tenantID) {
		return PreparedPlan{}, ErrInvalid
	}
	plan := command.Plan
	targetDefinitionSHA := plan.Target.DefinitionSHA256
	wire := planWire{
		ID: plan.ID.String(), ResourceKind: plan.Target.ResourceKind,
		SourceSchemaID: plan.SourceSchemaID.String(), TargetSchemaID: plan.Target.ID.String(),
		TargetVersion: plan.Target.Version, TargetDefinition: append([]byte(nil), plan.Target.DefinitionJSON...),
		TargetDefinitionSHA256: hex.EncodeToString(targetDefinitionSHA[:]),
		ConverterVersion:       plan.ConverterVersion, SourceCount: plan.SourceCount,
		PageSize: plan.PageSize, InitialRollingSHA256: hex.EncodeToString(plan.InitialRollingSHA256[:]),
		Mappings: make([]mappingWire, 0, len(plan.Mappings)), Pages: make([]pageWire, 0, len(plan.Pages)),
	}
	sourceAggregate := sha256.New()
	rolling := plan.InitialRollingSHA256
	for index, mapping := range plan.Mappings {
		field, _ := targetField(plan.Target.Fields, mapping.TargetFieldID)
		semanticSHA := field.SemanticSHA256
		wire.Mappings = append(wire.Mappings, mappingWire{index + 1, mapping.SourceFieldID.String(),
			mapping.TargetFieldID.String(), mapping.ConverterKind, hex.EncodeToString(semanticSHA[:])})
	}
	for index, page := range plan.Pages {
		reference := fmt.Sprintf("urn:harpy:iam:profile-backfill:%s:%s:%d:v1", tenantID, plan.ID, index+1)
		referenceSHA := sha256.Sum256([]byte(reference))
		var after *string
		if page.AfterResourceID != uuid.Nil {
			value := page.AfterResourceID.String()
			after = &value
		}
		wire.Pages = append(wire.Pages, pageWire{index + 1, page.JobID.String(),
			page.ScheduledRunAt.UTC().Format(time.RFC3339Nano), reference, hex.EncodeToString(referenceSHA[:]),
			after, page.ThroughResourceID.String(), page.ExpectedCount,
			hex.EncodeToString(page.SourceSHA256[:]), hex.EncodeToString(page.ExpectedTargetSHA256[:])})
		_, _ = sourceAggregate.Write(page.SourceSHA256[:])
		rolling = sha256.Sum256(append(append([]byte{}, rolling[:]...), page.ExpectedTargetSHA256[:]...))
	}
	wire.ExpectedSourceSHA256 = hex.EncodeToString(sourceAggregate.Sum(nil))
	wire.ExpectedTargetSHA256 = hex.EncodeToString(rolling[:])
	document, err := json.Marshal(wire)
	if err != nil || len(document) > 16<<20 {
		return PreparedPlan{}, ErrInvalid
	}
	digest := sha256.Sum256(document)
	first := wire.Pages[0]
	return PreparedPlan{Command: command, Document: document, PlanSHA256: digest, FirstJob: FirstJob{
		PlanID: plan.ID, JobID: plan.Pages[0].JobID, RunAt: plan.Pages[0].ScheduledRunAt.UTC(),
		PayloadReference: first.PayloadReference, PayloadSHA256: first.PayloadSHA,
		DedupeKey: fmt.Sprintf("iam.profile.backfill:%s:1", plan.ID),
	}}, nil
}
