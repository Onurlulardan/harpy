package adoption

func (state *validationState) validateExtensions() {
	state.validateProfileSchemas()
	fields := state.validateProfileFields()
	state.validateLayoutFields(fields)
	state.validateInvitations()
}
