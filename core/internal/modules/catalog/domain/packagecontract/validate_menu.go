package packagecontract

func validateMenu(
	wire menuWire,
	seen map[string]bool,
	components map[string]ComponentKind,
) (MenuItem, error) {
	id, err := ParseMenuItemID(wire.ID)
	if err != nil || seen[id.String()] || wire.Order < 0 || wire.Order > 100000 ||
		wire.Label == "" || len(wire.Label) > 128 {
		return MenuItem{}, ErrManifestInvalid
	}
	seen[id.String()] = true
	value := MenuItem{ID: id, Kind: wire.Kind, Label: wire.Label, Order: wire.Order}
	if wire.ParentID != "" {
		parent, err := ParseMenuItemID(wire.ParentID)
		if err != nil || parent.String() == id.String() {
			return MenuItem{}, ErrManifestInvalid
		}
		value.ParentID = &parent
	}
	if wire.Kind == MenuGroup {
		if wire.TargetID != "" || wire.URL != "" {
			return MenuItem{}, ErrManifestInvalid
		}
		return value, nil
	}
	if wire.Kind == MenuExternalLink {
		if wire.TargetID != "" || !validExternalURL(wire.URL) {
			return MenuItem{}, ErrManifestInvalid
		}
		value.URL = wire.URL
		return value, nil
	}
	target, err := ParseComponentID(wire.TargetID)
	if err != nil || wire.URL != "" || !menuMatchesComponent(wire.Kind, components[target.String()]) {
		return MenuItem{}, ErrManifestInvalid
	}
	value.TargetID = &target
	return value, nil
}

func menuMatchesComponent(menu MenuKind, component ComponentKind) bool {
	return menu == MenuForm && component == ComponentForm ||
		menu == MenuWorkflow && component == ComponentWorkflow ||
		menu == MenuDashboard && component == ComponentDashboard ||
		menu == MenuEndpoint && component == ComponentEndpoint
}

func menuParentsExist(items []MenuItem, identities map[string]bool) bool {
	for _, item := range items {
		if item.ParentID != nil && !identities[item.ParentID.String()] {
			return false
		}
	}
	return true
}
