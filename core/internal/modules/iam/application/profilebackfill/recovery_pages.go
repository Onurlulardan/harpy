package profilebackfill

import (
	"crypto/sha256"
	"fmt"
	"time"

	"github.com/google/uuid"
)

func BuildRecoveryPlan(input RecoveryPlanningInput) (RecoveryPlan, error) {
	if input.Parent.Validate() != nil || input.Parent.ExistingChildPlanID != uuid.Nil ||
		!stableID(input.ID) || input.ID == input.Parent.ID ||
		input.ID == recoveryRootID(input.Parent) || !input.Inventory.validated ||
		input.Inventory.parentPlanID != input.Parent.ID ||
		input.Inventory.sourceSchemaID != input.Parent.SourceSchemaID ||
		input.Inventory.targetSchemaID != input.Parent.Target.ID ||
		input.Inventory.pageSize != input.PageSize ||
		input.PageSize < 1 || input.PageSize > MaximumPageSize ||
		input.ScheduledRunAt.IsZero() || input.ScheduledRunAt.Location() != time.UTC {
		return RecoveryPlan{}, ErrInvalid
	}
	pages := materializeRecoveryPages(input.ID, input.ScheduledRunAt, input.Inventory.pages)
	plan := RecoveryPlan{
		id: input.ID, recoveryOfPlanID: input.Parent.ID,
		recoveryRootPlanID:    recoveryRootID(input.Parent),
		recoverySequence:      input.Parent.RecoverySequence + 1,
		parentPlanSHA256:      input.Parent.PlanSHA256,
		sourceSchemaID:        input.Parent.SourceSchemaID,
		target:                cloneTargetSchema(input.Parent.Target),
		converterVersion:      input.Parent.ConverterVersion,
		mappings:              cloneMappings(input.Parent.Mappings),
		inheritedTargetCount:  input.Inventory.inheritedTargetCount,
		inheritedTargetSHA256: input.Inventory.inheritedTargetSHA256,
		remainingSourceCount:  input.Inventory.remainingSourceCount,
		pageSize:              input.PageSize, initialRollingSHA256: sha256.Sum256(nil),
		pages: pages, validated: true,
	}
	if plan.Validate() != nil {
		return RecoveryPlan{}, ErrInvalid
	}
	return plan, nil
}

func PlanRecovery(
	id uuid.UUID,
	parent RecoveryParent,
	inventory RecoveryInventory,
	pageSize int,
	scheduledRunAt time.Time,
) (RecoveryPlan, error) {
	return BuildRecoveryPlan(RecoveryPlanningInput{
		ID: id, Parent: parent, Inventory: inventory,
		PageSize: pageSize, ScheduledRunAt: scheduledRunAt,
	})
}

func materializeRecoveryPages(
	planID uuid.UUID,
	runAt time.Time,
	evidence []recoveryPageEvidence,
) []Page {
	pages := make([]Page, 0, len(evidence))
	for index, item := range evidence {
		pageNumber := index + 1
		pages = append(pages, Page{
			JobID:                recoveryJobID(planID, pageNumber),
			ScheduledRunAt:       runAt.Add(time.Duration(index) * time.Nanosecond),
			AfterResourceID:      item.afterResourceID,
			ThroughResourceID:    item.throughResourceID,
			ExpectedCount:        item.expectedCount,
			SourceSHA256:         item.sourceSHA256,
			ExpectedTargetSHA256: item.expectedTargetSHA256,
		})
	}
	return pages
}

func recoveryRootID(parent RecoveryParent) uuid.UUID {
	if parent.RecoverySequence == 0 {
		return parent.ID
	}
	return parent.RecoveryRootPlanID
}

func recoveryJobID(planID uuid.UUID, pageNumber int) uuid.UUID {
	digest := sha256.Sum256([]byte(fmt.Sprintf("%s:%d", planID, pageNumber)))
	var result uuid.UUID
	copy(result[:], digest[:len(result)])
	result[6] = result[6]&0x0f | 0x70
	result[8] = result[8]&0x3f | 0x80
	return result
}
