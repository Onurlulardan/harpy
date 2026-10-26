package platformops

import (
	"net/http"

	"github.com/Onurlulardan/harpy/core/internal/platform/httpedge"
)

func (handler *HTTPHandler) Routes() []httpedge.Route {
	if handler == nil {
		return nil
	}
	return []httpedge.Route{
		{Method: http.MethodGet, Pattern: "/api/platform/operations/health", Handler: http.HandlerFunc(handler.getSystemHealth)},
		{Method: http.MethodGet, Pattern: "/api/platform/operations/traces", Handler: http.HandlerFunc(handler.listTraces)},
		{Method: http.MethodPost, Pattern: "/api/platform/operations/traces/search", Handler: http.HandlerFunc(handler.searchTraces)},
		{Method: http.MethodGet, Pattern: "/api/platform/operations/traces/stats", Handler: http.HandlerFunc(handler.getTraceStats)},
		{Method: http.MethodGet, Pattern: "/api/platform/operations/traces/spans/{span_id}", Handler: http.HandlerFunc(handler.getTraceSpan), Precedence: 1},
		{Method: http.MethodGet, Pattern: "/api/platform/operations/traces/{correlation_id}", Handler: http.HandlerFunc(handler.getTrace)},
		{Method: http.MethodGet, Pattern: "/api/platform/operations/traces/{correlation_id}/timeline", Handler: http.HandlerFunc(handler.getTraceTimeline)},
		{Method: http.MethodGet, Pattern: "/api/platform/operations/alert-rules", Handler: http.HandlerFunc(handler.listAlertRules)},
		{Method: http.MethodPost, Pattern: "/api/platform/operations/alert-rules", Handler: http.HandlerFunc(handler.createAlertRule)},
		{Method: http.MethodGet, Pattern: "/api/platform/operations/alert-rules/{alert_rule_id}", Handler: http.HandlerFunc(handler.getAlertRule)},
		{Method: http.MethodPut, Pattern: "/api/platform/operations/alert-rules/{alert_rule_id}", Handler: http.HandlerFunc(handler.updateAlertRule)},
		{Method: http.MethodDelete, Pattern: "/api/platform/operations/alert-rules/{alert_rule_id}", Handler: http.HandlerFunc(handler.deleteAlertRule)},
		{Method: http.MethodPost, Pattern: "/api/platform/operations/alert-rules/{alert_rule_id}/test", Handler: http.HandlerFunc(handler.testAlertRule)},
		{Method: http.MethodGet, Pattern: "/api/platform/operations/alert-events", Handler: http.HandlerFunc(handler.listAlertEvents)},
		{Method: http.MethodGet, Pattern: "/api/platform/operations/alert-events/{alert_event_id}", Handler: http.HandlerFunc(handler.getAlertEvent)},
		{Method: http.MethodPost, Pattern: "/api/platform/operations/permission-cache/flush", Handler: http.HandlerFunc(handler.flushPermissionCache)},
		{Method: http.MethodGet, Pattern: "/api/platform/database/stats", Handler: http.HandlerFunc(handler.getDatabaseStats)},
		{Method: http.MethodGet, Pattern: "/api/platform/database/schema", Handler: http.HandlerFunc(handler.getDatabaseSchema)},
		{Method: http.MethodGet, Pattern: "/api/platform/database/diagnostics/recovery-readiness", Handler: http.HandlerFunc(handler.getRecoveryReadiness)},
		{Method: http.MethodGet, Pattern: "/api/platform/recovery-protection", Handler: http.HandlerFunc(handler.getProtection)},
		{Method: http.MethodPost, Pattern: "/api/platform/recovery-protection/enable", Handler: http.HandlerFunc(handler.enableProtection)},
		{Method: http.MethodPost, Pattern: "/api/platform/recovery-protection/rotate", Handler: http.HandlerFunc(handler.rotateProtection)},
		{Method: http.MethodPost, Pattern: "/api/platform/recovery-protection/disable", Handler: http.HandlerFunc(handler.disableProtection)},
		{Method: http.MethodGet, Pattern: "/api/platform/recovery-protection/anchors", Handler: http.HandlerFunc(handler.listRecoveryAnchors)},
		{Method: http.MethodPost, Pattern: "/api/platform/recovery-protection/anchors/registration-options", Handler: http.HandlerFunc(handler.createRecoveryAnchorOptions)},
		{Method: http.MethodPost, Pattern: "/api/platform/recovery-protection/anchors", Handler: http.HandlerFunc(handler.registerRecoveryAnchor)},
		{Method: http.MethodDelete, Pattern: "/api/platform/recovery-protection/anchors/{anchor_id}", Handler: http.HandlerFunc(handler.deleteRecoveryAnchor)},
		{Method: http.MethodGet, Pattern: "/api/platform/recovery-operations/{operation_id}", Handler: http.HandlerFunc(handler.getRecoveryOperation)},
		{Method: http.MethodPost, Pattern: "/api/platform/recovery-operations/{operation_id}/kit-download", Handler: http.HandlerFunc(handler.downloadRecoveryKit)},
		{Method: http.MethodPost, Pattern: "/api/platform/recovery-operations/{operation_id}/kit-ack", Handler: http.HandlerFunc(handler.ackRecoveryKit)},
		{Method: http.MethodGet, Pattern: "/api/platform/recovery-sets", Handler: http.HandlerFunc(handler.listRecoverySets)},
		{Method: http.MethodPost, Pattern: "/api/platform/recovery-sets", Handler: http.HandlerFunc(handler.createRecoverySet)},
		{Method: http.MethodGet, Pattern: "/api/platform/recovery-sets/{recovery_set_id}", Handler: http.HandlerFunc(handler.getRecoverySet)},
		{Method: http.MethodDelete, Pattern: "/api/platform/recovery-sets/{recovery_set_id}", Handler: http.HandlerFunc(handler.deleteRecoverySet)},
		{Method: http.MethodPost, Pattern: "/api/platform/recovery-sets/{recovery_set_id}/manifest-download", Handler: http.HandlerFunc(handler.downloadRecoveryManifest)},
		{Method: http.MethodPost, Pattern: "/api/platform/recovery-sets/{recovery_set_id}/restore", Handler: http.HandlerFunc(handler.requestRestore)},
		{Method: http.MethodGet, Pattern: "/api/platform/restore-operations/{operation_id}", Handler: http.HandlerFunc(handler.getRestoreOperation)},
		{Method: http.MethodPost, Pattern: "/api/platform/restore-operations/{operation_id}/plan-download", Handler: http.HandlerFunc(handler.downloadRestorePlan)},
	}
}
