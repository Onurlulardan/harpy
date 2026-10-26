package systemtransfer

import "errors"

var (
	ErrExportWorkerInvalid         = errors.New("IAM system export worker input is invalid")
	ErrExportAttemptInvalid        = errors.New("IAM system export attempt is invalid")
	ErrExportAuthorizationRevoked  = errors.New("IAM system export authorization was revoked")
	ErrExportCancelled             = errors.New("IAM system export was cancelled")
	ErrExportLeaseLost             = errors.New("IAM system export lease was lost")
	ErrExportCapacity              = errors.New("IAM system export capacity is unavailable")
	ErrExportLimit                 = errors.New("IAM system export limit was exceeded")
	ErrExportQueryInvalid          = errors.New("IAM system export query is invalid")
	ErrExportCatalogPinStale       = errors.New("IAM system export catalog pin is stale")
	ErrExportArtifactTampered      = errors.New("IAM system export artifact evidence is invalid")
	ErrExportUnavailable           = errors.New("IAM system export dependency is unavailable")
	ErrExportStateConflict         = errors.New("IAM system export state conflicts")
	ErrExportOutcomeUnknown        = errors.New("IAM system export provider outcome is unknown")
	ErrExportOutcomeUnknownPersist = errors.New("IAM system export unknown outcome was not persisted")
	ErrExportReconciliationInvalid = errors.New("IAM system export reconciliation input is invalid")
	ErrExportReconciliationPersist = errors.New("IAM system export reconciliation was not persisted")
)
