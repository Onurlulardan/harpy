package asyncmanifestv1

import eventsv1 "github.com/Onurlulardan/harpy/contracts/events/v1"

const (
	WorkReminderDueEventType    = "work.reminder.due"
	WorkReminderDueSchemaSHA256 = "a70c6b913190e8de46e55e1b3d88a4d704e21a5c019b71adcd6b2071e9298827"
	WorkMutationSchemaSHA256    = "dd59769116d2848c51ffa702679533fc93642a4c277812e59aa236f8399a1f6b"
)

func WorkProducerPolicies() []eventsv1.ProducerPolicy {
	reminder, _ := eventsv1.NewProducerPolicy(
		WorkReminderDueEventType, 1, "work", WorkReminderDueSchemaSHA256,
	)
	policies := []eventsv1.ProducerPolicy{reminder}
	for _, eventType := range workMutationEventTypes() {
		policy, _ := eventsv1.NewProducerPolicy(
			eventType, 1, "work", WorkMutationSchemaSHA256)
		policies = append(policies, policy)
	}
	return policies
}

func workMutationEventTypes() []string {
	return []string{
		"work.board.archive", "work.board.create", "work.board.restore",
		"work.board.update", "work.calendar.change", "work.checklist.change",
		"work.column.create", "work.column.update", "work.comment.create",
		"work.comment.redact", "work.label.change", "work.membership.change",
		"work.projection.apply", "work.projection.repair", "work.recurrence.change",
		"work.recurrence.materialize", "work.task.archive", "work.task.assign",
		"work.task.create", "work.task.move", "work.task.restore",
		"work.task.transition", "work.task.update",
	}
}
