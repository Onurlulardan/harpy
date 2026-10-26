package profilebackfill

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"

	profileapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/profile"
	profilebackfillapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/profilebackfill"
	"github.com/google/uuid"
)

type RecoveryContract struct {
	Parent              profilebackfillapp.RecoveryParent
	Source              profileapp.Schema
	Target              profileapp.Schema
	ExistingChildPlanID uuid.UUID
}

type recoveryParentRow struct {
	ID                  uuid.UUID
	RecoveryRootPlanID  *uuid.UUID
	RecoverySequence    int64
	PlanSHA256          []byte
	ResourceKind        string
	SourceSchemaID      uuid.UUID
	TargetSchemaID      uuid.UUID
	ConverterVersion    string
	ExistingChildPlanID *uuid.UUID
	SourceSnapshot      []byte
	TargetSnapshot      []byte
	Mappings            []byte
}

func decodeRecoveryContract(row recoveryParentRow) (RecoveryContract, error) {
	source, err := decodeRecoverySchema(row.SourceSnapshot)
	if err != nil {
		return RecoveryContract{}, err
	}
	target, err := decodeRecoverySchema(row.TargetSnapshot)
	if err != nil {
		return RecoveryContract{}, err
	}
	mappings, err := decodeRecoveryMappings(target, row.Mappings)
	if err != nil || len(row.PlanSHA256) != sha256.Size ||
		source.ID() != row.SourceSchemaID || target.ID() != row.TargetSchemaID ||
		source.State() != profileapp.SchemaActive ||
		target.State() != profileapp.SchemaActive ||
		source.Resource() != target.Resource() ||
		string(source.Resource()) != row.ResourceKind ||
		target.Version() != source.Version()+1 {
		return RecoveryContract{}, profilebackfillapp.ErrUnavailable
	}
	parent := profilebackfillapp.RecoveryParent{
		ID: row.ID, RecoverySequence: row.RecoverySequence,
		State:            profilebackfillapp.RecoveryParentFenced,
		SourceSchemaID:   row.SourceSchemaID,
		ConverterVersion: row.ConverterVersion, Mappings: mappings,
		Target: recoveryTargetSchema(target),
	}
	copy(parent.PlanSHA256[:], row.PlanSHA256)
	clear(row.PlanSHA256)
	if row.RecoveryRootPlanID != nil {
		parent.RecoveryRootPlanID = *row.RecoveryRootPlanID
	}
	if row.ExistingChildPlanID != nil {
		parent.ExistingChildPlanID = *row.ExistingChildPlanID
	}
	if parent.Validate() != nil {
		return RecoveryContract{}, profilebackfillapp.ErrUnavailable
	}
	return RecoveryContract{
		Parent: parent, Source: source, Target: target,
		ExistingChildPlanID: parent.ExistingChildPlanID,
	}, nil
}

func recoveryTargetSchema(schema profileapp.Schema) profilebackfillapp.TargetSchema {
	fields := make([]profilebackfillapp.TargetField, 0, len(schema.Fields()))
	for _, field := range schema.Fields() {
		fields = append(fields, profilebackfillapp.TargetField{
			ID: field.ID(), Required: field.Required(),
			Active:         field.State() == profileapp.FieldActive,
			SemanticSHA256: field.SemanticSHA256(),
		})
	}
	return profilebackfillapp.TargetSchema{
		ID: schema.ID(), ResourceKind: string(schema.Resource()),
		Version: schema.Version(), DefinitionJSON: schema.DefinitionJSON(),
		DefinitionSHA256: schema.DefinitionSHA256(), Fields: fields,
	}
}

func decodeRecoveryMappings(
	target profileapp.Schema,
	document []byte,
) ([]profilebackfillapp.Mapping, error) {
	var wires []recoveryMappingWire
	if len(document) == 0 || json.Unmarshal(document, &wires) != nil ||
		len(wires) == 0 || len(wires) > 500 {
		return nil, profilebackfillapp.ErrUnavailable
	}
	result := make([]profilebackfillapp.Mapping, 0, len(wires))
	for _, wire := range wires {
		sourceID, sourceErr := uuid.Parse(wire.SourceFieldID)
		targetID, targetErr := uuid.Parse(wire.TargetFieldID)
		semantic, semanticErr := hex.DecodeString(wire.SemanticSHA)
		field, found := target.Field(targetID)
		if sourceErr != nil || targetErr != nil || semanticErr != nil ||
			sourceID == uuid.Nil || targetID == uuid.Nil || !found ||
			len(semantic) != sha256.Size ||
			(wire.ConverterKind != profilebackfillapp.MappingCopy &&
				wire.ConverterKind != profilebackfillapp.MappingNormalize) {
			return nil, profilebackfillapp.ErrUnavailable
		}
		expected := field.SemanticSHA256()
		if !hmac.Equal(semantic, expected[:]) {
			return nil, profilebackfillapp.ErrUnavailable
		}
		result = append(result, profilebackfillapp.Mapping{
			SourceFieldID: sourceID, TargetFieldID: targetID,
			ConverterKind: wire.ConverterKind,
		})
	}
	return result, nil
}
