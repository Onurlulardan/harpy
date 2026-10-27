package profilepostgres

import (
	"crypto/hmac"
	"encoding/hex"
	"encoding/json"

	profileapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/profile"
	"github.com/google/uuid"
)

const profileWriteContractV1 = "harpy/iam-profile-write-contract/v1"

func decodeProfileWriteContract(payload []byte) (profileWriteContract, error) {
	var wire profileWriteContractWire
	if len(payload) == 0 || json.Unmarshal(payload, &wire) != nil ||
		wire.Contract != profileWriteContractV1 || len(wire.Source) == 0 ||
		wire.Mappings == nil {
		return profileWriteContract{}, errProfile
	}
	source, err := decodeProfileSchema(wire.Source)
	if err != nil {
		return profileWriteContract{}, errProfile
	}
	result := profileWriteContract{source: source}
	if wire.PlanID == nil {
		if len(wire.Mappings) != 0 || string(wire.Target) != "null" {
			return profileWriteContract{}, errProfile
		}
		return result, nil
	}
	result.planID, err = uuid.Parse(*wire.PlanID)
	if err != nil || result.planID == uuid.Nil || len(wire.Target) == 0 {
		return profileWriteContract{}, errProfile
	}
	result.target, err = decodeProfileSchema(wire.Target)
	if err != nil || len(wire.Mappings) == 0 ||
		len(wire.Mappings) > profileapp.MaximumActiveFields {
		return profileWriteContract{}, errProfile
	}
	result.mappings = make([]profileapp.MigrationMapping, len(wire.Mappings))
	for index, mapping := range wire.Mappings {
		compiled, compileErr := decodeProfileWriteMapping(result.target, mapping)
		if compileErr != nil {
			return profileWriteContract{}, errProfile
		}
		result.mappings[index] = compiled
	}
	return result, nil
}

func decodeProfileWriteMapping(
	target profileapp.Schema,
	wire profileWriteMappingWire,
) (profileapp.MigrationMapping, error) {
	sourceID, sourceErr := uuid.Parse(wire.SourceFieldID)
	targetID, targetErr := uuid.Parse(wire.TargetFieldID)
	semantic, semanticErr := hex.DecodeString(wire.SemanticSHA)
	converter := profileapp.MigrationConverter(wire.Converter)
	field, found := target.Field(targetID)
	if sourceErr != nil || targetErr != nil || semanticErr != nil ||
		sourceID == uuid.Nil || targetID == uuid.Nil || !converter.Valid() ||
		!found || len(semantic) != 32 {
		return profileapp.MigrationMapping{}, errProfile
	}
	expected := field.SemanticSHA256()
	if !hmac.Equal(semantic, expected[:]) {
		return profileapp.MigrationMapping{}, errProfile
	}
	return profileapp.MigrationMapping{
		SourceFieldID: sourceID, TargetFieldID: targetID, Converter: converter,
	}, nil
}
