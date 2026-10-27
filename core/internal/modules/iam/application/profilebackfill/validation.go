package profilebackfill

import (
	"bytes"
	"crypto/sha256"
	"time"

	"github.com/google/uuid"
)

func (evidence Evidence) Validate() error {
	if !stableID(evidence.OperationID) || evidence.RequestSHA256 == [32]byte{} ||
		!stableID(evidence.ActorID) || !stableID(evidence.ManagePermissionID) ||
		evidence.ExpectedTokenVersion < 1 || evidence.ExpectedAuthorizationVersion < 1 {
		return ErrInvalid
	}
	return nil
}

func (plan Plan) Validate() error {
	if !stableID(plan.ID) || !stableID(plan.SourceSchemaID) ||
		plan.Target.ID == plan.SourceSchemaID || !stableID(plan.Target.ID) ||
		(plan.Target.ResourceKind != "user" && plan.Target.ResourceKind != "organization" &&
			plan.Target.ResourceKind != "role") || plan.Target.Version < 2 ||
		len(plan.Target.DefinitionJSON) < 2 || len(plan.Target.DefinitionJSON) > 4<<20 ||
		sha256.Sum256(plan.Target.DefinitionJSON) != plan.Target.DefinitionSHA256 ||
		plan.ConverterVersion != ConverterTypedMap || plan.SourceCount < 0 ||
		plan.PageSize < 1 || plan.PageSize > 250 || plan.InitialRollingSHA256 == [32]byte{} ||
		len(plan.Mappings) < 1 || len(plan.Mappings) > 500 ||
		len(plan.Pages) < 1 || len(plan.Pages) > 65536 {
		return ErrInvalid
	}
	if !validMappings(plan.Target, plan.Mappings) || !validPages(plan) {
		return ErrInvalid
	}
	return nil
}

func validMappings(target TargetSchema, mappings []Mapping) bool {
	sources, targets := map[uuid.UUID]bool{}, map[uuid.UUID]bool{}
	for _, mapping := range mappings {
		field, found := targetField(target.Fields, mapping.TargetFieldID)
		if !stableID(mapping.SourceFieldID) || !stableID(mapping.TargetFieldID) || !found ||
			field.SemanticSHA256 == [32]byte{} ||
			(mapping.ConverterKind != MappingCopy && mapping.ConverterKind != MappingNormalize) ||
			sources[mapping.SourceFieldID] || targets[mapping.TargetFieldID] ||
			(mapping.SourceFieldID == mapping.TargetFieldID && mapping.ConverterKind != MappingCopy) {
			return false
		}
		sources[mapping.SourceFieldID], targets[mapping.TargetFieldID] = true, true
	}
	for _, field := range target.Fields {
		if !stableID(field.ID) || field.SemanticSHA256 == [32]byte{} ||
			field.Active && field.Required && !targets[field.ID] {
			return false
		}
	}
	return true
}

func targetField(fields []TargetField, id uuid.UUID) (TargetField, bool) {
	for _, field := range fields {
		if field.ID == id {
			return field, true
		}
	}
	return TargetField{}, false
}

func validPages(plan Plan) bool {
	jobs := map[uuid.UUID]bool{}
	var total int64
	var previous uuid.UUID
	rolling := plan.InitialRollingSHA256
	for index, page := range plan.Pages {
		if !stableID(page.JobID) || page.JobID.Version() != uuid.Version(7) || jobs[page.JobID] ||
			page.ScheduledRunAt.IsZero() || page.ScheduledRunAt.Location() != time.UTC ||
			!stableID(page.ThroughResourceID) || page.ExpectedCount < 0 ||
			page.ExpectedCount > plan.PageSize || page.SourceSHA256 == [32]byte{} ||
			page.ExpectedTargetSHA256 == [32]byte{} {
			return false
		}
		if index == 0 && page.AfterResourceID != uuid.Nil ||
			index > 0 && page.AfterResourceID != previous ||
			page.AfterResourceID != uuid.Nil && bytes.Compare(page.AfterResourceID[:], page.ThroughResourceID[:]) >= 0 {
			return false
		}
		jobs[page.JobID], previous = true, page.ThroughResourceID
		total += int64(page.ExpectedCount)
		rolling = sha256.Sum256(append(append([]byte{}, rolling[:]...), page.ExpectedTargetSHA256[:]...))
	}
	return total == plan.SourceCount && rolling != [32]byte{}
}
