package api

import (
	"context"

	platformopsmodule "github.com/Onurlulardan/harpy/core/internal/modules/platformops"
	"github.com/Onurlulardan/harpy/core/internal/platform/httpedge"
)

type platformOpsHTTPRuntime struct {
	runtime *platformopsmodule.Runtime
}

func (runtime *platformOpsHTTPRuntime) Routes() []httpedge.Route {
	return runtime.runtime.Routes()
}

func (runtime *platformOpsHTTPRuntime) Observe(
	ctx context.Context,
	request httpedge.RequestObservation,
) error {
	return runtime.runtime.Observe(ctx, request)
}
