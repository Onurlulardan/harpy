package api

import (
	"fmt"
	"time"

	notificationsmodule "github.com/Onurlulardan/harpy/core/internal/modules/notifications"
	platformasync "github.com/Onurlulardan/harpy/core/internal/platform/async"
	"github.com/Onurlulardan/harpy/core/internal/platform/config"
	"github.com/Onurlulardan/harpy/core/internal/platform/httpedge"
	"github.com/Onurlulardan/harpy/core/internal/platform/postgres"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type notificationsHTTPRuntime struct {
	handler     *notificationsmodule.HTTPHandler
	application *notificationsmodule.HTTPRuntime
	workload    *notificationsProjectorWorkload
}

func buildNotificationsHTTP(settings config.API, pool *postgres.Pool,
	catalogs platformasync.WriteCatalogs, iamAccess notificationsIAMAccess,
	report httpedge.ErrorSink,
) (notificationsHTTPRuntime, error) {
	if pool == nil || iamAccess == nil || report == nil {
		return notificationsHTTPRuntime{},
			fmt.Errorf("Notifications HTTP composition dependencies are required")
	}
	coordinator, err := platformtx.NewCoordinator(pool)
	if err != nil {
		return notificationsHTTPRuntime{}, fmt.Errorf("build Notifications coordinator: %w", err)
	}
	keyring, err := settings.NotificationsConsumerKeyring()
	if err != nil {
		return notificationsHTTPRuntime{}, fmt.Errorf("load Notifications consumer keys: %w", err)
	}
	access, err := newNotificationsAccessVerifier(iamAccess,
		notificationConsumerVerifier{keyring: keyring, now: time.Now})
	if err != nil {
		return notificationsHTTPRuntime{}, err
	}
	application, err := notificationsmodule.NewPlatformHTTPRuntime(
		notificationsmodule.PlatformHTTPRuntimeDependencies{
			Generations: pool, Scopes: coordinator, Catalogs: catalogs})
	if err != nil {
		return notificationsHTTPRuntime{}, fmt.Errorf("build Notifications HTTP application: %w", err)
	}
	hub := notificationsmodule.NewRealtimeHub()
	handler, err := notificationsmodule.NewHTTPHandler(notificationsmodule.HTTPHandlerDependencies{
		Access: access, Application: application, Realtime: hub,
		Report: notificationsmodule.HTTPErrorSink(report)})
	if err != nil {
		return notificationsHTTPRuntime{}, fmt.Errorf("build Notifications HTTP handler: %w", err)
	}
	if !handler.Ready() {
		return notificationsHTTPRuntime{}, fmt.Errorf("build Notifications HTTP handler: not ready")
	}
	projector, err := application.NewRealtimeProjectionRuntime(hub)
	if err != nil {
		return notificationsHTTPRuntime{}, fmt.Errorf("build Notifications projector: %w", err)
	}
	workload, err := newNotificationsProjectorWorkload(projector, hub)
	if err != nil {
		return notificationsHTTPRuntime{}, fmt.Errorf("build Notifications projector workload: %w", err)
	}
	return notificationsHTTPRuntime{
		handler: handler, application: application, workload: workload,
	}, nil
}

func (runtime notificationsHTTPRuntime) Ready() bool {
	return runtime.handler != nil && runtime.handler.Ready() &&
		runtime.application != nil && runtime.workload != nil
}

func (runtime notificationsHTTPRuntime) Routes() []httpedge.Route {
	if !runtime.Ready() {
		return nil
	}
	return runtime.handler.Routes()
}
