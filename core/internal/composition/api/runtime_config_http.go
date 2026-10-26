package api

import (
	"fmt"
	"net/http"

	"github.com/Onurlulardan/harpy/core/internal/platform/buildinfo"
	"github.com/Onurlulardan/harpy/core/internal/platform/config"
	"github.com/Onurlulardan/harpy/core/internal/platform/httpedge"
	"github.com/Onurlulardan/harpy/core/internal/platform/httpwire"
)

type runtimeConfigSnapshot struct {
	deploymentMode    config.DeploymentMode
	primaryTenantSlug string
	aiChatComposed    bool
	workComposed      bool
}

type runtimeConfigHandler struct {
	response runtimeConfigEnvelope
	report   httpedge.ErrorSink
}

type runtimeConfigEnvelope struct {
	Success bool              `json:"success"`
	Data    runtimeConfigData `json:"data"`
}

type runtimeConfigData struct {
	DeploymentMode      string                           `json:"deployment_mode"`
	BackendVersion      string                           `json:"backend_version"`
	DefaultTenantSlug   *string                          `json:"default_tenant_slug"`
	ServiceAvailability runtimeConfigServiceAvailability `json:"service_availability"`
}

type runtimeConfigServiceAvailability struct {
	AIChat bool `json:"ai_chat"`
	Work   bool `json:"work"`
}

func newRuntimeConfigHandler(
	info buildinfo.Info,
	snapshot runtimeConfigSnapshot,
	report httpedge.ErrorSink,
) (*runtimeConfigHandler, error) {
	version := info.Version()
	if err := version.Validate(); err != nil || report == nil {
		return nil, fmt.Errorf("runtime config HTTP dependencies are invalid")
	}
	if err := snapshot.deploymentMode.Validate(); err != nil {
		return nil, err
	}
	var slug *string
	switch snapshot.deploymentMode {
	case config.DeploymentOnPrem:
		if snapshot.primaryTenantSlug == "" {
			return nil, fmt.Errorf("on-premise primary tenant slug is required")
		}
		value := snapshot.primaryTenantSlug
		slug = &value
	case config.DeploymentSaaS:
		if snapshot.primaryTenantSlug != "" {
			return nil, fmt.Errorf("SaaS runtime config cannot expose the control tenant slug")
		}
	}
	return &runtimeConfigHandler{
		response: runtimeConfigEnvelope{Success: true, Data: runtimeConfigData{
			DeploymentMode: string(snapshot.deploymentMode),
			BackendVersion: version.Version, DefaultTenantSlug: slug,
			ServiceAvailability: runtimeConfigServiceAvailability{
				AIChat: snapshot.aiChatComposed,
				Work:   snapshot.workComposed,
			},
		}},
		report: report,
	}, nil
}

func (handler *runtimeConfigHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if handler == nil || handler.report == nil || r == nil {
		return
	}
	err := httpwire.WriteJSON(
		w, http.StatusOK, layoutHTTPCorrelationID(r), handler.response,
		r.Method == http.MethodHead,
	)
	if err != nil {
		handler.report(fmt.Errorf("write runtime config response: %w", err))
	}
}

func runtimeConfigRoutes(handler *runtimeConfigHandler) ([]httpedge.Route, error) {
	if handler == nil {
		return nil, fmt.Errorf("runtime config HTTP handler is required")
	}
	return []httpedge.Route{
		{Method: http.MethodGet, Pattern: "/api/app/config", Handler: handler},
		{Method: http.MethodHead, Pattern: "/api/app/config", Handler: handler},
	}, nil
}
