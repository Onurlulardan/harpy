package asyncmanifestv1

import eventsv1 "github.com/Onurlulardan/harpy/contracts/events/v1"

const (
	WorkflowInstanceChangedEventType    = "workflow.instance.changed"
	WorkflowInstanceChangedSchemaSHA256 = "8db0032c583d8961f702aea39f47ea1ac7455530c8e2049347c6e48a63428342"
)

func WorkflowProducerPolicies() []eventsv1.ProducerPolicy {
	policy, _ := eventsv1.NewProducerPolicy(
		WorkflowInstanceChangedEventType,
		1,
		"workflow",
		WorkflowInstanceChangedSchemaSHA256,
	)
	return []eventsv1.ProducerPolicy{policy}
}
