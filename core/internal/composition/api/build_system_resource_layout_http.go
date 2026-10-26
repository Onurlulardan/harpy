package api

import (
	"fmt"

	iammodule "github.com/Onurlulardan/harpy/core/internal/modules/iam"
	layouts "github.com/Onurlulardan/harpy/core/internal/orchestration/systemresourcelayouts"
	"github.com/Onurlulardan/harpy/core/internal/platform/httpedge"
)

func buildSystemResourceLayoutHTTP(
	application *layouts.Application,
	access *iammodule.SystemLayoutAccessRuntime,
	report httpedge.ErrorSink,
) (*systemResourceLayoutHTTPHandler, error) {
	if application == nil || access == nil || report == nil {
		return nil, fmt.Errorf("system resource layout HTTP dependencies are required")
	}
	handler, err := newSystemResourceLayoutHTTPHandler(systemResourceLayoutHTTPDependencies{
		Access: access, Application: application,
		Report: systemResourceLayoutHTTPErrorSink(report),
	})
	if err != nil {
		return nil, fmt.Errorf("build system resource layout HTTP handler: %w", err)
	}
	if !systemResourceLayoutHTTPReady(handler) {
		return nil, fmt.Errorf("system resource layout HTTP surface is incomplete")
	}
	return handler, nil
}

func systemResourceLayoutHTTPReady(handler *systemResourceLayoutHTTPHandler) bool {
	return handler != nil && len(handler.Routes()) == 7
}
