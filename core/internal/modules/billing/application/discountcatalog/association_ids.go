package discountcatalog

import "github.com/google/uuid"

func nextAssociationIDs(next IDGenerator) ([4]uuid.UUID, error) {
	var result [4]uuid.UUID
	for index := range result {
		identifier, err := nextV7(next)
		if err != nil {
			return [4]uuid.UUID{}, err
		}
		result[index] = identifier
	}
	return result, nil
}

func validCampaignUpdate(value CampaignUpdate) bool {
	return !value.Replace || value.CampaignID == nil || validV7(*value.CampaignID)
}
