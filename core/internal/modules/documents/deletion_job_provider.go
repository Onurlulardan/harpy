package documents

import (
	"time"

	"github.com/Onurlulardan/harpy/core/internal/platform/async/effectexecution"
)

type deletionProvider struct {
	identity effectexecution.Identity
	effect   DeletionEffect
	now      func() time.Time
}

func newDeletionProvider(effect DeletionEffect,
	now func() time.Time,
) (*deletionProvider, error) {
	identity, err := effectexecution.NewIdentity("documents.blob", "delete")
	if err != nil || nilArchivePort(effect) || now == nil {
		return nil, ErrDocumentsRuntime
	}
	return &deletionProvider{identity: identity, effect: effect, now: now}, nil
}

func (provider *deletionProvider) Identity() effectexecution.Identity {
	if provider == nil {
		return effectexecution.Identity{}
	}
	return provider.identity
}

var _ effectexecution.Provider = (*deletionProvider)(nil)
