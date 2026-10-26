package application

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/platformauthority/api"
	"github.com/Onurlulardan/harpy/core/internal/modules/platformauthority/ports"
)

func (service *Coordinator) FenceSigner(
	ctx context.Context,
	request api.SignerFenceRequest,
) error {
	if request.RegistryBindingID == "" || request.ExpectedSignerGeneration < 1 ||
		request.ExpectedSignerKID == "" || !validDigest(request.EvidenceSHA256) {
		return api.ErrConflict
	}
	binding, err := service.store.LoadBinding(ctx, request.RegistryBindingID)
	if err != nil {
		return mapStoreError(err)
	}
	if binding.SignerGeneration != request.ExpectedSignerGeneration ||
		binding.SignerKID != request.ExpectedSignerKID {
		return api.ErrStaleHead
	}
	err = service.store.FenceSigner(ctx, ports.SignerFenceRecord{
		BindingID:                request.RegistryBindingID,
		ExpectedSignerGeneration: request.ExpectedSignerGeneration,
		ExpectedSignerKID:        request.ExpectedSignerKID,
		EvidenceSHA256:           request.EvidenceSHA256,
	})
	return mapOptionalStoreError(err)
}

func mapOptionalStoreError(err error) error {
	if err == nil {
		return nil
	}
	return mapStoreError(err)
}

var _ api.SignerLifecycle = (*Coordinator)(nil)
