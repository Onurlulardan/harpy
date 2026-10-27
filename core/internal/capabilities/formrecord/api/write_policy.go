package recordkernel

type WritePolicy string

const (
	WriteActor         WritePolicy = "actor"
	WriteActorWorkflow WritePolicy = "actor_workflow"
	WriteWorkflow      WritePolicy = "workflow"
	WriteEngine        WritePolicy = "engine"
)

func (policy WritePolicy) valid() bool {
	switch policy {
	case WriteActor, WriteActorWorkflow, WriteWorkflow, WriteEngine:
		return true
	default:
		return false
	}
}

func (policy WritePolicy) allowsActor() bool {
	return policy == WriteActor || policy == WriteActorWorkflow
}

func (policy WritePolicy) allowsWorkflow() bool {
	return policy == WriteWorkflow || policy == WriteActorWorkflow
}
