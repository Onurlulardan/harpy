package providerregistry

import (
	"encoding/json"
	"fmt"

	canonicaljsonv1 "github.com/Onurlulardan/harpy/contracts/canonicaljson/v1"
	"github.com/Onurlulardan/harpy/core/internal/modules/platformauthority/ports"
	"github.com/google/uuid"
)

func canonicalProviderBinding(binding ports.ProviderBinding) ([]byte, error) {
	if _, err := uuid.Parse(binding.BindingID); err != nil {
		return nil, fmt.Errorf("SaaS authority provider binding is invalid")
	}
	if _, err := uuid.Parse(binding.EnvironmentID); err != nil ||
		binding.SignerGeneration < 1 || binding.SignerKID == "" ||
		len(binding.SignerPublicKeySHA256) != 64 {
		return nil, fmt.Errorf("SaaS authority provider binding is invalid")
	}
	document, err := json.Marshal(binding)
	if err != nil {
		return nil, fmt.Errorf("encode SaaS authority provider binding")
	}
	canonical, err := canonicaljsonv1.CanonicalizeObject(document, 16*1024)
	if err != nil {
		return nil, fmt.Errorf("canonicalize SaaS authority provider binding")
	}
	return canonical, nil
}
