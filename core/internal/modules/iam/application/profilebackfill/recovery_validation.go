package profilebackfill

import (
	"bytes"
	"crypto/sha256"
	"time"

	"github.com/google/uuid"
)

func (parent RecoveryParent) Validate() error {
	rootValid := parent.RecoverySequence == 0 && parent.RecoveryRootPlanID == uuid.Nil ||
		parent.RecoverySequence > 0 && stableID(parent.RecoveryRootPlanID) &&
			parent.RecoveryRootPlanID != parent.ID
	childValid := parent.ExistingChildPlanID == uuid.Nil ||
		stableID(parent.ExistingChildPlanID)
	if !stableID(parent.ID) || !rootValid || !childValid ||
		parent.State != RecoveryParentFenced ||
		parent.PlanSHA256 == [sha256.Size]byte{} ||
		!validRecoveryContract(parent.SourceSchemaID, parent.Target,
			parent.ConverterVersion, parent.Mappings) {
		return ErrInvalid
	}
	return nil
}

func (plan RecoveryPlan) Validate() error {
	if !plan.validated || !stableID(plan.id) || !stableID(plan.recoveryOfPlanID) ||
		!stableID(plan.recoveryRootPlanID) || plan.recoverySequence < 1 ||
		plan.id == plan.recoveryOfPlanID || plan.id == plan.recoveryRootPlanID ||
		plan.parentPlanSHA256 == [sha256.Size]byte{} ||
		!validRecoveryContract(plan.sourceSchemaID, plan.target,
			plan.converterVersion, plan.mappings) ||
		plan.inheritedTargetCount < 0 ||
		plan.inheritedTargetSHA256 == [sha256.Size]byte{} ||
		plan.remainingSourceCount < 0 || plan.pageSize < 1 ||
		plan.pageSize > MaximumPageSize ||
		plan.initialRollingSHA256 == [sha256.Size]byte{} ||
		!validRecoveryPages(plan) {
		return ErrInvalid
	}
	return nil
}

func validRecoveryContract(
	sourceID uuid.UUID,
	target TargetSchema,
	converter string,
	mappings []Mapping,
) bool {
	if !stableID(sourceID) || !stableID(target.ID) || target.ID == sourceID ||
		(target.ResourceKind != "user" && target.ResourceKind != "organization" &&
			target.ResourceKind != "role") || target.Version < 2 ||
		len(target.DefinitionJSON) < 2 || len(target.DefinitionJSON) > 4<<20 ||
		sha256.Sum256(target.DefinitionJSON) != target.DefinitionSHA256 ||
		converter != ConverterTypedMap || len(mappings) < 1 || len(mappings) > 500 {
		return false
	}
	return validMappings(target, mappings)
}

func validRecoveryPages(plan RecoveryPlan) bool {
	if plan.remainingSourceCount == 0 {
		return len(plan.pages) == 0
	}
	if len(plan.pages) == 0 || len(plan.pages) > 65536 {
		return false
	}
	jobs := make(map[uuid.UUID]struct{}, len(plan.pages))
	var total int64
	var previous uuid.UUID
	for index, page := range plan.pages {
		if !validRecoveryPage(page, plan.pageSize) {
			return false
		}
		if _, duplicate := jobs[page.JobID]; duplicate {
			return false
		}
		if index == 0 && page.AfterResourceID != uuid.Nil ||
			index > 0 && page.AfterResourceID != previous ||
			page.AfterResourceID != uuid.Nil &&
				bytes.Compare(page.AfterResourceID[:], page.ThroughResourceID[:]) >= 0 {
			return false
		}
		jobs[page.JobID] = struct{}{}
		previous = page.ThroughResourceID
		total += int64(page.ExpectedCount)
	}
	return total == plan.remainingSourceCount
}

func validRecoveryPage(page Page, pageSize int) bool {
	return stableID(page.JobID) && page.JobID.Version() == uuid.Version(7) &&
		!page.ScheduledRunAt.IsZero() && page.ScheduledRunAt.Location() == time.UTC &&
		stableID(page.ThroughResourceID) && page.ExpectedCount > 0 &&
		page.ExpectedCount <= pageSize && page.SourceSHA256 != [sha256.Size]byte{} &&
		page.ExpectedTargetSHA256 != [sha256.Size]byte{}
}

func cloneMappings(mappings []Mapping) []Mapping {
	return append([]Mapping(nil), mappings...)
}

func cloneTargetSchema(target TargetSchema) TargetSchema {
	target.DefinitionJSON = bytes.Clone(target.DefinitionJSON)
	target.Fields = append([]TargetField(nil), target.Fields...)
	return target
}
