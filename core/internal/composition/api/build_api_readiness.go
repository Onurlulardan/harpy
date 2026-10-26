package api

import documentsmodule "github.com/Onurlulardan/harpy/core/internal/modules/documents"

func apiRuntimeReadiness(
	iam iamAPIRuntime,
	analytics analyticsHTTPRuntime,
	catalog catalogHTTPRuntime,
	documents *documentsmodule.HTTPHandler,
	notifications notificationsHTTPRuntime,
	work workHTTPRuntime,
	secondary secondaryHTTPRuntime,
	layouts *systemResourceLayoutHTTPHandler,
) []func() bool {
	return []func() bool{
		iam.actorBridgeReady,
		analytics.Ready,
		catalog.Ready,
		documents.Ready,
		notifications.Ready,
		work.Ready,
		secondary.Ready,
		func() bool { return systemResourceLayoutHTTPReady(layouts) },
	}
}
