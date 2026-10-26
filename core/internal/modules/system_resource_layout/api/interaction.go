package api

type Interaction string

const (
	InteractionAuto     Interaction = "auto"
	InteractionReadOnly Interaction = "read_only"
	InteractionDisabled Interaction = "disabled"
)

func (interaction Interaction) Valid() bool {
	return interaction == InteractionAuto || interaction == InteractionReadOnly ||
		interaction == InteractionDisabled
}
