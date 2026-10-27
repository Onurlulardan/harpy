package notifications

import (
	"net/http"

	notificationhttp "github.com/Onurlulardan/harpy/core/internal/modules/notifications/adapters/http"
	"github.com/Onurlulardan/harpy/core/internal/platform/httpedge"
	"github.com/Onurlulardan/harpy/core/internal/platform/requestmeta"
	"github.com/google/uuid"
)

type HTTPErrorSink = notificationhttp.ErrorSink
type HTTPApplication = notificationhttp.Application
type HTTPAccess = notificationhttp.Access
type HTTPAccessRequest = notificationhttp.AccessRequest
type HTTPAccessVerifier = notificationhttp.AccessVerifier
type HTTPPermission = notificationhttp.Permission
type HTTPRealtimeTransport = notificationhttp.RealtimeTransport
type HTTPRealtimeHint = notificationhttp.RealtimeHint
type HTTPHandlerDependencies struct {
	Access      HTTPAccessVerifier
	Application HTTPApplication
	Realtime    HTTPRealtimeTransport
	Report      HTTPErrorSink
}

type HTTPHandler struct{ transport *notificationhttp.Handler }

const (
	PermissionInboxRead       = notificationhttp.PermissionInboxRead
	PermissionInboxUpdate     = notificationhttp.PermissionInboxUpdate
	PermissionInboxDelete     = notificationhttp.PermissionInboxDelete
	PermissionPreferencesRead = notificationhttp.PermissionPreferencesRead
	PermissionPreferencesEdit = notificationhttp.PermissionPreferencesEdit
	PermissionSend            = notificationhttp.PermissionSend
)

const (
	OperationInboxList       = notificationhttp.OperationInboxList
	OperationInboxGet        = notificationhttp.OperationInboxGet
	OperationInboxRead       = notificationhttp.OperationInboxRead
	OperationInboxReadAll    = notificationhttp.OperationInboxReadAll
	OperationInboxDismiss    = notificationhttp.OperationInboxDismiss
	OperationPreferencesGet  = notificationhttp.OperationPreferencesGet
	OperationPreferencesEdit = notificationhttp.OperationPreferencesEdit
	OperationIntentCreate    = notificationhttp.OperationIntentCreate
	OperationTicketCreate    = notificationhttp.OperationTicketCreate
	OperationSocketConnect   = notificationhttp.OperationSocketConnect
	OperationLegacyConnect   = notificationhttp.OperationLegacyConnect
)

var (
	ErrHTTPAuthentication = notificationhttp.ErrAuthentication
	ErrHTTPForbidden      = notificationhttp.ErrForbidden
	ErrHTTPInvalid        = notificationhttp.ErrInvalid
	ErrHTTPNotFound       = notificationhttp.ErrNotFound
	ErrHTTPConflict       = notificationhttp.ErrConflict
	ErrHTTPUnavailable    = notificationhttp.ErrUnavailable
)

func NewHTTPAccess(principalID, tenantID uuid.UUID, tokenVersion,
	authorizationVersion int64, permission HTTPPermission, operation, origin string,
	sessionAuthenticated, csrfVerified bool,
) (HTTPAccess, error) {
	return notificationhttp.NewAccess(principalID, tenantID, tokenVersion,
		authorizationVersion, permission, operation, origin,
		sessionAuthenticated, csrfVerified)
}

func NewLegacyHTTPAccess(principalID, tenantID uuid.UUID, tokenVersion,
	authorizationVersion int64, origin string, manifestDigest [32]byte,
	sessionAuthenticated bool,
) (HTTPAccess, error) {
	return notificationhttp.NewLegacyAccess(principalID, tenantID, tokenVersion,
		authorizationVersion, origin, manifestDigest, sessionAuthenticated)
}

func NewHTTPHandler(dependencies HTTPHandlerDependencies) (*HTTPHandler, error) {
	transport, err := notificationhttp.NewHandler(notificationhttp.HandlerDependencies{
		Access: dependencies.Access, Application: dependencies.Application,
		Realtime: dependencies.Realtime, Report: dependencies.Report,
		Correlation: notificationCorrelation})
	if err != nil {
		return nil, err
	}
	return &HTTPHandler{transport: transport}, nil
}

func (handler *HTTPHandler) Ready() bool {
	return handler != nil && handler.transport != nil && handler.transport.Ready()
}

func (handler *HTTPHandler) Routes() []httpedge.Route {
	if !handler.Ready() {
		return nil
	}
	routes := handler.transport.Routes()
	result := make([]httpedge.Route, len(routes))
	for index, route := range routes {
		result[index] = httpedge.Route{Method: route.Method,
			Pattern: route.Pattern, Handler: route.Handler}
	}
	return result
}

func NewRealtimeHub() *notificationhttp.Hub { return notificationhttp.NewHub() }

func notificationCorrelation(r *http.Request) string {
	if r != nil {
		if value, found := requestmeta.CorrelationIDFromContext(r.Context()); found {
			return value.String()
		}
	}
	return requestmeta.EntropyFailureCorrelationID.String()
}
