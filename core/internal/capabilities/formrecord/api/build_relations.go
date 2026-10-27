package recordkernel

import (
	"github.com/google/uuid"
)

func relationProfiles(
	schema StorageSchema,
) (map[uuid.UUID]RelationProfile, error) {
	result := make(map[uuid.UUID]RelationProfile, len(schema.Relations))
	for _, relation := range schema.Relations {
		fieldID, err := parseStableID(relation.FieldID)
		relationID, relationErr := parseStableID(relation.ID)
		targetFormID, targetErr := parseStableID(relation.TargetFormID)
		if err != nil || relationErr != nil || targetErr != nil ||
			relation.Cardinality != "one" && relation.Cardinality != "many" ||
			relation.DeletePolicy != "restrict" && relation.DeletePolicy != "cascade" {
			return nil, ErrCatalogInvalid
		}
		if _, duplicate := result[fieldID]; duplicate {
			return nil, ErrCatalogInvalid
		}
		result[fieldID] = RelationProfile{RelationID: relationID,
			TargetFormID: targetFormID, Cardinality: relation.Cardinality,
			DeletePolicy: relation.DeletePolicy}
	}
	return result, nil
}
