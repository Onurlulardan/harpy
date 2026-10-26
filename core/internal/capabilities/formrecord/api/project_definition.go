package recordkernel

import (
	"encoding/json"
	"fmt"
	"slices"

	canonicaljsonv1 "github.com/Onurlulardan/harpy/contracts/canonicaljson/v1"
	"github.com/google/uuid"
)

func ProjectDefinition(document []byte) (Projection, error) {
	envelope, err := decodeDefinition(document)
	if err != nil {
		return Projection{}, fmt.Errorf("%w: canonical definition", ErrDefinitionInvalid)
	}
	var schema StorageSchema
	if strictDecode(envelope.schema, &schema) != nil {
		return Projection{}, ErrDefinitionInvalid
	}
	var rules RecordRules
	switch envelope.contract {
	case DefinitionV1Contract:
		rules = legacyRecordRules(schema)
	case DefinitionV2Contract:
		if envelope.recordContract != RecordRulesContract ||
			strictDecode(envelope.record, &rules) != nil {
			return Projection{}, ErrDefinitionInvalid
		}
	default:
		return Projection{}, ErrDefinitionInvalid
	}
	projected, err := envelope.v1Projection()
	if err != nil {
		return Projection{}, ErrDefinitionInvalid
	}
	recordBytes, err := canonicalObject(rules, MaximumRecordBytes)
	if err != nil {
		return Projection{}, ErrDefinitionInvalid
	}
	projection, err := buildProjection(envelope.contract, envelope.canonical,
		projected, rules, recordBytes)
	if err != nil {
		return Projection{}, ErrDefinitionInvalid
	}
	targets, err := relationTargetFormIDs(schema.Relations)
	if err != nil {
		return Projection{}, ErrDefinitionInvalid
	}
	projection.relationTargetFormIDs = targets
	return projection, nil
}

func relationTargetFormIDs(relations []RelationDefinition) ([]uuid.UUID, error) {
	seen := make(map[uuid.UUID]struct{}, len(relations))
	result := make([]uuid.UUID, 0, len(relations))
	for _, relation := range relations {
		target, err := uuid.Parse(relation.TargetFormID)
		if err != nil || target == uuid.Nil || target.Version() != 7 ||
			target.Variant() != uuid.RFC4122 || target.String() != relation.TargetFormID {
			return nil, ErrDefinitionInvalid
		}
		if _, found := seen[target]; found {
			continue
		}
		seen[target] = struct{}{}
		result = append(result, target)
	}
	slices.SortFunc(result, func(left, right uuid.UUID) int {
		return slices.Compare(left[:], right[:])
	})
	return result, nil
}

func canonicalObject(value any, maximum int) ([]byte, error) {
	encoded, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	return canonicaljsonv1.CanonicalizeObject(encoded, maximum)
}

func legacyRecordRules(schema StorageSchema) RecordRules {
	fields := make([]RecordFieldRule, len(schema.Fields))
	for index, field := range schema.Fields {
		policy := WriteActor
		if field.Role == RoleCalculation {
			policy = WriteEngine
		}
		fields[index] = RecordFieldRule{FieldID: field.ID, WritePolicy: policy}
	}
	slices.SortFunc(fields, func(left, right RecordFieldRule) int {
		return compareText(left.FieldID, right.FieldID)
	})
	return RecordRules{Fields: fields, WorkflowActions: []string{},
		Defaults: []DefaultRule{}, CrossFieldValidations: []CrossFieldValidation{}}
}

func compareText(left, right string) int {
	if left < right {
		return -1
	}
	if left > right {
		return 1
	}
	return 0
}
