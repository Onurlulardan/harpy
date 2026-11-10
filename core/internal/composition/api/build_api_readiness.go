package api

import documentsmodule "github.com/Onurlulardan/harpy/core/internal/modules/documents"

func apiRuntimeReadiness(
	iam iamAPIRuntime,
	documents *documentsmodule.HTTPHandler,
	notifications notificationsHTTPRuntime,
	secondary secondaryHTTPRuntime,
) []func() bool {
	return []func() bool{
		iam.actorBridgeReady,
		documents.Ready,
		notifications.Ready,
		secondary.Ready,
	}
}
