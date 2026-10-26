package documents

import (
	"time"

	"github.com/Onurlulardan/harpy/core/internal/platform/async/effectexecution"
)

type adoptionCopyProvider struct {
	identity effectexecution.Identity
	effect   AdoptionCopyEffect
	now      func() time.Time
}

func newAdoptionCopyProvider(effect AdoptionCopyEffect,
	now func() time.Time,
) (*adoptionCopyProvider, error) {
	identity, err := effectexecution.NewIdentity("documents.adoption", "copy")
	if err != nil || nilArchivePort(effect) || now == nil {
		return nil, ErrDocumentsRuntime
	}
	return &adoptionCopyProvider{identity: identity, effect: effect, now: now}, nil
}

func (provider *adoptionCopyProvider) Identity() effectexecution.Identity {
	if provider == nil {
		return effectexecution.Identity{}
	}
	return provider.identity
}

var _ effectexecution.Provider = (*adoptionCopyProvider)(nil)
