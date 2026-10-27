package analytics

import (
	"context"
	"errors"
	"fmt"
	"time"

	effectsv1 "github.com/Onurlulardan/harpy/contracts/effects/v1"
	exportapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/dashboardexport"
	"github.com/Onurlulardan/harpy/core/internal/platform/async/effectexecution"
	"github.com/google/uuid"
)

type dashboardExportJobProvider struct {
	identity  effectexecution.Identity
	kind      string
	operation string
	worker    dashboardExportJobWorker
	now       func() time.Time
}

func newDashboardExportJobProvider(kind, operation string,
	worker dashboardExportJobWorker, now func() time.Time,
) (*dashboardExportJobProvider, error) {
	identity, err := effectexecution.NewIdentity(dashboardExportEffectProvider, operation)
	if err != nil || worker == nil || now == nil ||
		!((kind == DashboardExportDeleteJobKind && operation == "delete") ||
			(kind == DashboardExportExpireJobKind && operation == "expire")) {
		return nil, exportapp.ErrUnavailable
	}
	return &dashboardExportJobProvider{identity: identity, kind: kind,
		operation: operation, worker: worker, now: now}, nil
}

func (provider *dashboardExportJobProvider) Identity() effectexecution.Identity {
	if provider == nil {
		return effectexecution.Identity{}
	}
	return provider.identity
}

func (provider *dashboardExportJobProvider) Execute(ctx context.Context,
	request effectsv1.Request,
) effectexecution.Result {
	tenantID, exportID, jobID, leaseOwner, err := provider.validateRequest(request)
	if err != nil || ctx == nil || ctx.Err() != nil {
		return effectexecution.Unknown()
	}
	var result exportapp.RenderResult
	if provider.operation == "expire" {
		result, err = provider.worker.ProcessExpiry(
			ctx, tenantID, exportID, jobID, leaseOwner)
	} else {
		result, err = provider.worker.ProcessDelete(
			ctx, tenantID, exportID, jobID, leaseOwner)
	}
	if errors.Is(err, exportapp.ErrOutcomeUnknown) || ctx.Err() != nil {
		return effectexecution.Unknown()
	}
	if err != nil {
		code, class := dashboardExportFailure(err)
		return provider.failed(request, exportID, code, class)
	}
	if dashboardExportJobSucceeded(provider.operation, result.State) {
		return provider.succeeded(request, exportID, result.State)
	}
	if result.State == exportapp.StateQueued || result.State == exportapp.StateRunning ||
		result.State == exportapp.StateDeleting {
		return provider.failed(request, exportID, "ANALYTICS_EXPORT_RETRY",
			effectsv1.ErrorTransient)
	}
	return provider.failed(request, exportID, "ANALYTICS_EXPORT_TERMINAL",
		effectsv1.ErrorPermanent)
}

func (provider *dashboardExportJobProvider) validateRequest(request effectsv1.Request) (
	uuid.UUID, uuid.UUID, uuid.UUID, uuid.UUID, error,
) {
	if provider == nil || request.Validate() != nil || request.TenantID == nil ||
		request.Authority.ScopeKind != effectsv1.ScopeTenant ||
		request.Authority.Domain != "analytics" || request.Authority.WriterOrigin != effectsv1.WriterV2 ||
		request.Provider != provider.identity.Provider() ||
		request.Operation != provider.operation || request.EffectKind != provider.kind ||
		request.EffectRevision != 1 {
		return uuid.Nil, uuid.Nil, uuid.Nil, uuid.Nil,
			fmt.Errorf("Analytics dashboard export effect request is invalid")
	}
	tenantID, tenantErr := uuid.Parse(*request.TenantID)
	exportID, exportErr := uuid.Parse(request.PayloadReference)
	jobID, jobErr := uuid.Parse(request.CausationID)
	leaseOwner, leaseErr := uuid.Parse(request.EffectID)
	digest := dashboardExportReferenceDigest(request.PayloadReference)
	if tenantErr != nil || exportErr != nil || jobErr != nil || leaseErr != nil ||
		tenantID == uuid.Nil || exportID.Version() != uuid.Version(7) ||
		jobID.Version() != uuid.Version(7) || leaseOwner.Version() != uuid.Version(5) ||
		leaseOwner.Variant() != uuid.RFC4122 ||
		request.PayloadSHA256 != digest || request.IdempotencyKey !=
		provider.kind+":"+request.PayloadReference {
		return uuid.Nil, uuid.Nil, uuid.Nil, uuid.Nil,
			fmt.Errorf("Analytics dashboard export effect evidence is invalid")
	}
	return tenantID, exportID, jobID, leaseOwner, nil
}

func dashboardExportFailure(err error) (string, effectsv1.ErrorClass) {
	if errors.Is(err, exportapp.ErrForbidden) || errors.Is(err, exportapp.ErrInvalid) ||
		errors.Is(err, exportapp.ErrNotFound) {
		return "ANALYTICS_EXPORT_REJECTED", effectsv1.ErrorPermanent
	}
	if errors.Is(err, exportapp.ErrConflict) {
		return "ANALYTICS_EXPORT_CONFLICT", effectsv1.ErrorConflict
	}
	return "ANALYTICS_EXPORT_UNAVAILABLE", effectsv1.ErrorTransient
}

func dashboardExportJobSucceeded(operation string, state exportapp.State) bool {
	return state == exportapp.StateDeleted || operation == "expire" &&
		(state == exportapp.StateCancelled || state == exportapp.StateExpired ||
			state == exportapp.StateFailed)
}
