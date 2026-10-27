package profilebackfill

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

func PrepareRecovery(
	command RecoveryCreateCommand,
	tenantID uuid.UUID,
) (PreparedRecoveryPlan, error) {
	if command.Evidence.Validate() != nil || command.Plan.Validate() != nil ||
		!stableID(tenantID) {
		return PreparedRecoveryPlan{}, ErrInvalid
	}
	plan := command.Plan
	target := plan.target
	wire := recoveryPlanWire{
		ID: plan.id.String(), ResourceKind: target.ResourceKind,
		RecoveryOfPlanID:   plan.recoveryOfPlanID.String(),
		RecoveryRootPlanID: plan.recoveryRootPlanID.String(),
		RecoverySequence:   plan.recoverySequence,
		ParentPlanSHA256:   hex.EncodeToString(plan.parentPlanSHA256[:]),
		SourceSchemaID:     plan.sourceSchemaID.String(), TargetSchemaID: target.ID.String(),
		TargetVersion:          target.Version,
		TargetDefinitionSHA256: hex.EncodeToString(target.DefinitionSHA256[:]),
		ConverterVersion:       plan.converterVersion,
		InheritedTargetCount:   plan.inheritedTargetCount,
		InheritedTargetSHA256:  hex.EncodeToString(plan.inheritedTargetSHA256[:]),
		RemainingSourceCount:   plan.remainingSourceCount, PageSize: plan.pageSize,
		InitialRollingSHA256: hex.EncodeToString(plan.initialRollingSHA256[:]),
		Mappings:             make([]mappingWire, 0, len(plan.mappings)),
		Pages:                make([]pageWire, 0, len(plan.pages)),
	}
	for index, mapping := range plan.mappings {
		field, _ := targetField(target.Fields, mapping.TargetFieldID)
		wire.Mappings = append(wire.Mappings, mappingWire{
			Ordinal: index + 1, SourceFieldID: mapping.SourceFieldID.String(),
			TargetFieldID: mapping.TargetFieldID.String(),
			ConverterKind: mapping.ConverterKind,
			SemanticSHA:   hex.EncodeToString(field.SemanticSHA256[:]),
		})
	}
	remainingSourceSHA := sha256.Sum256(nil)
	rolling := plan.initialRollingSHA256
	for index, page := range plan.pages {
		reference := fmt.Sprintf(
			"urn:harpy:iam:profile-backfill:%s:%s:%d:v1",
			tenantID, plan.id, index+1,
		)
		referenceSHA := sha256.Sum256([]byte(reference))
		var after *string
		if page.AfterResourceID != uuid.Nil {
			value := page.AfterResourceID.String()
			after = &value
		}
		wire.Pages = append(wire.Pages, pageWire{
			PageNumber: index + 1, JobID: page.JobID.String(),
			ScheduledRunAt:   page.ScheduledRunAt.UTC().Format(time.RFC3339Nano),
			PayloadReference: reference,
			PayloadSHA:       hex.EncodeToString(referenceSHA[:]),
			AfterResourceID:  after, ThroughResourceID: page.ThroughResourceID.String(),
			ExpectedCount:     page.ExpectedCount,
			SourceSHA:         hex.EncodeToString(page.SourceSHA256[:]),
			ExpectedTargetSHA: hex.EncodeToString(page.ExpectedTargetSHA256[:]),
		})
		remainingSourceSHA = sha256.Sum256(append(
			append([]byte{}, remainingSourceSHA[:]...), page.SourceSHA256[:]...,
		))
		rolling = sha256.Sum256(append(
			append([]byte{}, rolling[:]...), page.ExpectedTargetSHA256[:]...,
		))
	}
	wire.RemainingSourceSHA256 = hex.EncodeToString(remainingSourceSHA[:])
	wire.ExpectedRemainingTargetSHA256 = hex.EncodeToString(rolling[:])
	document, err := json.Marshal(wire)
	if err != nil || len(document) > 16<<20 {
		return PreparedRecoveryPlan{}, ErrInvalid
	}
	prepared := PreparedRecoveryPlan{
		Command: command, Document: document, PlanSHA256: sha256.Sum256(document),
	}
	if len(plan.pages) > 0 {
		first := wire.Pages[0]
		prepared.FirstJob = &FirstJob{
			PlanID: plan.id, JobID: plan.pages[0].JobID,
			RunAt:            plan.pages[0].ScheduledRunAt.UTC(),
			PayloadReference: first.PayloadReference, PayloadSHA256: first.PayloadSHA,
			DedupeKey: fmt.Sprintf("iam.profile.backfill:%s:1", plan.id),
		}
	}
	return prepared, nil
}
