package recordkernel

func compileWorkflowActions(values []string) (map[Action]struct{}, error) {
	result := make(map[Action]struct{}, len(values))
	previous := ""
	for _, value := range values {
		action := Action(value)
		if value <= previous || action != ActionCreate &&
			action != ActionUpdate && action != ActionDelete {
			return nil, ErrCatalogInvalid
		}
		previous = value
		result[action] = struct{}{}
	}
	return result, nil
}
