package api

import (
	"fmt"

	iammodule "github.com/Onurlulardan/harpy/core/internal/modules/iam"
	systemresources "github.com/Onurlulardan/harpy/core/internal/orchestration/systemresources"
	"github.com/Onurlulardan/harpy/core/internal/platform/httpedge"
)

func buildSystemResourceRecordHTTP(
	application *systemresources.Application,
	access *iammodule.SystemLayoutAccessRuntime,
	report httpedge.ErrorSink,
) (*systemResourceRecordHTTPHandler, error) {
	if application == nil || access == nil || report == nil {
		return nil, fmt.Errorf("system resource record HTTP dependencies are required")
	}
	canonical, err := newCanonicalSystemResourceRecordHTTPApplication(application)
	if err != nil {
		return nil, fmt.Errorf("build canonical system resource HTTP application: %w", err)
	}
	handler, err := newSystemResourceRecordHTTPHandler(
		systemResourceRecordHTTPDependencies{
			Access: access, Application: canonical,
			Report: systemResourceRecordHTTPErrorSink(report),
		},
	)
	if err != nil {
		return nil, fmt.Errorf("build system resource record HTTP handler: %w", err)
	}
	if !systemResourceRecordHTTPReady(handler) {
		return nil, fmt.Errorf("system resource record HTTP surface is incomplete")
	}
	return handler, nil
}

func systemResourceRecordHTTPReady(handler *systemResourceRecordHTTPHandler) bool {
	return handler != nil && len(handler.Routes()) == 15
}
