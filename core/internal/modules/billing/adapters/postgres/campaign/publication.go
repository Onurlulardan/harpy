package campaign

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"

	campaigndomain "github.com/Onurlulardan/harpy/core/internal/modules/billing/domain/campaign"
)

func publicationDocument(publication campaigndomain.Publication) (string, [32]byte, error) {
	document, err := json.Marshal(struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		BannerText  string `json:"banner_text"`
		AutoApply   bool   `json:"auto_apply"`
	}{Name: publication.Name(), Description: publication.Description(),
		BannerText: publication.BannerText(), AutoApply: publication.AutoApply()})
	if err != nil {
		return "", [32]byte{}, fmt.Errorf("encode billing campaign presentation: %w", err)
	}
	return string(document), sha256.Sum256(document), nil
}
