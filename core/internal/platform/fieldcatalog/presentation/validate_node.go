package presentation

func (state *validationState) validateNode(
	node Node,
	path string,
	parentVariant string,
	depth int,
	seenFields map[string]struct{},
) error {
	state.nodeCount++
	if depth > maximumDepth || state.nodeCount > maximumNodes ||
		!stableUUIDv7(node.ID) || node.Span < 1 || node.Span > 4 {
		return invalid(path, "node identity, depth or budget")
	}
	if _, duplicate := state.nodeIDs[node.ID]; duplicate {
		return invalid(path+".id", "duplicate")
	}
	state.nodeIDs[node.ID] = struct{}{}
	if err := state.validateVisibility(path+".visibility", node.Visibility); err != nil {
		return err
	}
	switch node.Kind {
	case "container":
		return state.validateContainer(node, path, parentVariant, depth, seenFields)
	case "field":
		return state.validateFieldNode(node, path, seenFields)
	case "text":
		return state.validateTextNode(node, path)
	case "divider":
		return state.validateDividerNode(node, path)
	default:
		return invalid(path+".kind", "unknown")
	}
}

func (state *validationState) validateContainer(
	node Node, path string, parentVariant string, depth int,
	seenFields map[string]struct{},
) error {
	if node.FieldID != "" || node.Widget != "" || node.Interaction != "" ||
		node.Children == nil || !validContainerVariant(node.Variant) {
		return invalid(path, "container shape")
	}
	if node.Variant == "grid" {
		if node.Columns < 1 || node.Columns > 4 {
			return invalid(path+".columns", "grid columns")
		}
	} else if node.Columns != 0 {
		return invalid(path+".columns", "only grid declares columns")
	}
	labelled := node.Variant == "section" || node.Variant == "card" || node.Variant == "tab"
	if err := state.validateTranslations(path+".translations", node.Translations, labelled); err != nil {
		return err
	}
	if !labelled && len(node.Translations) != 0 {
		return invalid(path+".translations", "unlabelled container")
	}
	if node.Variant == "tab" && parentVariant != "tabs" ||
		node.Variant != "tab" && parentVariant == "tabs" {
		return invalid(path+".variant", "tabs topology")
	}
	for index, child := range node.Children {
		if err := state.validateNode(
			child, path+".children["+decimal(index)+"]", node.Variant, depth+1, seenFields,
		); err != nil {
			return err
		}
	}
	return nil
}

func (state *validationState) validateFieldNode(
	node Node, path string, seenFields map[string]struct{},
) error {
	field, exists := state.fields[node.FieldID]
	modeAllowed := exists && (field.mode == FieldEditable || node.Interaction != "auto")
	if !modeAllowed || node.Variant != "" || node.Columns != 0 || node.Children != nil ||
		!validInteraction(node.Interaction) || !state.registry.accepts(node.Widget, field.typeID) {
		return invalid(path, "field shape, mode or widget compatibility")
	}
	if _, duplicate := seenFields[node.FieldID]; duplicate {
		return invalid(path+".field_id", "duplicate in surface")
	}
	seenFields[node.FieldID] = struct{}{}
	return state.validateTranslations(path+".translations", node.Translations, true)
}

func (state *validationState) validateTextNode(node Node, path string) error {
	if node.FieldID != "" || node.Widget != "" || node.Interaction != "" ||
		node.Columns != 0 || node.Children != nil || !validTextVariant(node.Variant) {
		return invalid(path, "text shape")
	}
	return state.validateTranslations(path+".translations", node.Translations, true)
}

func (state *validationState) validateDividerNode(node Node, path string) error {
	if node.FieldID != "" || node.Widget != "" || node.Interaction != "" ||
		node.Variant != "" || node.Columns != 0 || node.Children != nil ||
		node.Visibility != nil || node.Translations == nil || len(node.Translations) != 0 {
		return invalid(path, "divider shape")
	}
	return nil
}
