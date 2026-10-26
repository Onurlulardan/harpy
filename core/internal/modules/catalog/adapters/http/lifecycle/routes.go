package lifecyclehttp

func (handler *Handler) Routes() []Route {
	if !handler.Ready() {
		return nil
	}
	routes := []Route{
		handler.ListInstallationsRoute(), handler.InstallRoute(),
		handler.GetInstallationRoute(), handler.UpgradeRoute(),
		handler.DisableRoute(), handler.EnableRoute(), handler.RestoreRoute(),
		handler.RetireRoute(), handler.ApprovePurgeRoute(), handler.PurgeRoute(),
		handler.ReadOperationRoute(), handler.CancelOperationRoute(),
		handler.RetryOperationRoute(),
	}
	if handler.demo != nil {
		routes = append(routes, handler.GetDemoStatusRoute(), handler.InstallDemoRoute(),
			handler.PlanDemoRemovalRoute(), handler.RemoveDemoRoute())
	}
	return routes
}
