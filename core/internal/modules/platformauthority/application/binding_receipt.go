package application

import (
	"github.com/Onurlulardan/harpy/core/internal/modules/platformauthority/api"
	"github.com/Onurlulardan/harpy/core/internal/modules/platformauthority/ports"
)

func bindingReceipt(record ports.BindingRecord) api.BindingReceipt {
	return api.BindingReceipt{
		BindingID: record.BindingID, EnvironmentID: record.EnvironmentID,
		RegistryKind: string(record.RegistryKind), SignerGeneration: record.SignerGeneration,
		SignerKID:             record.SignerKID,
		SignerPublicKeySHA256: record.SignerPublicKeySHA256,
	}
}
