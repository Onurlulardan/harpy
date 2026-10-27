package documents

import (
	"time"

	"github.com/Onurlulardan/harpy/core/internal/platform/async/effectexecution"
)

type uploadReconcileProvider struct {
	identity effectexecution.Identity
	effect   UploadReconcileEffect
	now      func() time.Time
}

func newUploadReconcileProvider(effect UploadReconcileEffect,
	now func() time.Time,
) (*uploadReconcileProvider, error) {
	identity, err := effectexecution.NewIdentity("documents.upload", "reconcile")
	if err != nil || nilArchivePort(effect) || now == nil {
		return nil, ErrDocumentsRuntime
	}
	return &uploadReconcileProvider{identity: identity, effect: effect, now: now}, nil
}

func (provider *uploadReconcileProvider) Identity() effectexecution.Identity {
	if provider == nil {
		return effectexecution.Identity{}
	}
	return provider.identity
}

var _ effectexecution.Provider = (*uploadReconcileProvider)(nil)
