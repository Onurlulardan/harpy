package authorization

import "fmt"

type Snapshot struct {
	actor        *ActorFact
	resource     *ResourceFact
	action       *ActionFact
	clock        int64
	bindings     []BindingFact
	implications []ImplicationFact
	valid        bool
}

func NewSnapshot(
	actor *ActorFact,
	resource *ResourceFact,
	action *ActionFact,
	clock int64,
	bindings []BindingFact,
	implications []ImplicationFact,
) (Snapshot, error) {
	snapshot := Snapshot{
		actor: cloneActor(actor), resource: cloneResource(resource),
		action: cloneAction(action), clock: clock,
		bindings:     append([]BindingFact(nil), bindings...),
		implications: append([]ImplicationFact(nil), implications...), valid: true,
	}
	if err := snapshot.Validate(); err != nil {
		return Snapshot{}, err
	}
	return snapshot, nil
}

func (snapshot Snapshot) Validate() error {
	if !snapshot.valid || snapshot.clock < 0 ||
		(snapshot.actor != nil && !snapshot.actor.valid()) ||
		(snapshot.resource != nil && !snapshot.resource.valid()) ||
		(snapshot.action != nil && !snapshot.action.valid()) {
		return fmt.Errorf("IAM authorization snapshot is invalid")
	}
	for _, fact := range snapshot.bindings {
		if !fact.valid() {
			return fmt.Errorf("IAM authorization binding evidence is invalid")
		}
	}
	for _, fact := range snapshot.implications {
		if !fact.valid() {
			return fmt.Errorf("IAM authorization implication evidence is invalid")
		}
	}
	return nil
}

func cloneActor(value *ActorFact) *ActorFact {
	if value == nil {
		return nil
	}
	result := *value
	return &result
}

func cloneResource(value *ResourceFact) *ResourceFact {
	if value == nil {
		return nil
	}
	result := *value
	return &result
}

func cloneAction(value *ActionFact) *ActionFact {
	if value == nil {
		return nil
	}
	result := *value
	return &result
}
