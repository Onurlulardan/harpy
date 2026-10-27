package api

import (
	"context"
	"fmt"

	usagemodule "github.com/Onurlulardan/harpy/core/internal/modules/usage"
	"github.com/Onurlulardan/harpy/core/internal/platform/httpedge"
	platformpostgres "github.com/Onurlulardan/harpy/core/internal/platform/postgres"
)

type usageHTTPRuntime struct {
	handler  *usagemodule.HTTPHandler
	recorder interface {
		Record(context.Context, usagemodule.RecordRequest) (usagemodule.RecordResult, error)
	}
}

func (runtime usageHTTPRuntime) Ready() bool {
	return runtime.handler != nil && runtime.handler.Ready() && runtime.recorder != nil
}

func (runtime usageHTTPRuntime) Routes() []httpedge.Route {
	if !runtime.Ready() {
		return nil
	}
	return runtime.handler.Routes()
}

func (runtime usageHTTPRuntime) Record(
	ctx context.Context, request usagemodule.RecordRequest,
) (usagemodule.RecordResult, error) {
	if !runtime.Ready() {
		return usagemodule.RecordResult{}, usagemodule.ErrHTTPUnavailable
	}
	return runtime.recorder.Record(ctx, request)
}

func buildUsageHTTP(
	pool *platformpostgres.Pool,
	access usagemodule.HTTPAccessVerifier,
	report httpedge.ErrorSink,
) (usageHTTPRuntime, error) {
	profile, err := newProfileQueryRuntime(pool)
	if err != nil {
		return usageHTTPRuntime{}, err
	}
	store, err := usagemodule.NewPostgresStoreForProfile(
		profile.usageRunner(), platformpostgres.RowNotFound,
	)
	if err != nil {
		return usageHTTPRuntime{}, fmt.Errorf("build Usage store: %w", err)
	}
	application, err := usagemodule.NewService(
		store, moduleRuntimeClock{}, usageRuntimeIDs{},
	)
	if err != nil {
		return usageHTTPRuntime{}, fmt.Errorf("build Usage application: %w", err)
	}
	handler, err := usagemodule.NewHTTPHandler(usagemodule.HTTPHandlerDependencies{
		Access: access, Application: application,
		Report: usagemodule.HTTPErrorSink(report),
	})
	if err != nil {
		return usageHTTPRuntime{}, fmt.Errorf("build Usage HTTP: %w", err)
	}
	if !handler.Ready() || len(handler.Routes()) != 16 {
		return usageHTTPRuntime{}, fmt.Errorf("Usage HTTP route contract is incomplete")
	}
	runtime := usageHTTPRuntime{handler: handler, recorder: application}
	if !runtime.Ready() {
		return usageHTTPRuntime{}, fmt.Errorf("Usage HTTP runtime is incomplete")
	}
	return runtime, nil
}
