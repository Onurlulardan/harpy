package campaign

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"

	campaigndomain "github.com/Onurlulardan/harpy/core/internal/modules/billing/domain/campaign"
)

type digestPublication struct {
	Name               string   `json:"name"`
	Description        string   `json:"description"`
	BannerText         string   `json:"banner_text"`
	AutoApply          bool     `json:"auto_apply"`
	Priority           int32    `json:"priority"`
	StartsAt           int64    `json:"starts_at"`
	EndsAt             *int64   `json:"ends_at"`
	DiscountVersionIDs []string `json:"discount_version_ids"`
}

func publicationDigest(publication campaigndomain.Publication) digestPublication {
	identifiers := publication.DiscountVersionIDs()
	discounts := make([]string, len(identifiers))
	for index, identifier := range identifiers {
		discounts[index] = hex.EncodeToString(identifier[:])
	}
	var end *int64
	if value := publication.EndsAt(); value != nil {
		unixMicro := value.UnixMicro()
		end = &unixMicro
	}
	return digestPublication{Name: publication.Name(), Description: publication.Description(),
		BannerText: publication.BannerText(), AutoApply: publication.AutoApply(),
		Priority: publication.Priority(), StartsAt: publication.StartsAt().UnixMicro(),
		EndsAt: end, DiscountVersionIDs: discounts}
}

func requestDigest(value any) ([32]byte, error) {
	document, err := json.Marshal(value)
	if err != nil {
		return [32]byte{}, err
	}
	return sha256.Sum256(document), nil
}
