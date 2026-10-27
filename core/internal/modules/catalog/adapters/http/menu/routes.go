package menuhttp

func (handler *Handler) Routes() []Route {
	if !handler.Ready() {
		return nil
	}
	return []Route{
		handler.GetMenuRoute(),
		handler.ReplaceMenuRoute(),
		handler.UpdateMenuOverrideRoute(),
		handler.GetActiveMenuConfigRoute(),
		handler.CreateMenuConfigRoute(),
		handler.UpdateMenuConfigRoute(),
	}
}
