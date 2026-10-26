package documents

import (
	"time"

	"github.com/Onurlulardan/harpy/core/internal/platform/async/effectexecution"
)

type archiveProvider struct {
	identity effectexecution.Identity
	effect   ArchiveEffect
	now      func() time.Time
}

func newArchiveProvider(effect ArchiveEffect,
	now func() time.Time,
) (*archiveProvider, error) {
	identity, err := effectexecution.NewIdentity("documents.archive", "build")
	if err != nil || nilArchivePort(effect) || now == nil {
		return nil, ErrDocumentsRuntime
	}
	return &archiveProvider{identity: identity, effect: effect, now: now}, nil
}

func (provider *archiveProvider) Identity() effectexecution.Identity {
	if provider == nil {
		return effectexecution.Identity{}
	}
	return provider.identity
}

var _ effectexecution.Provider = (*archiveProvider)(nil)
