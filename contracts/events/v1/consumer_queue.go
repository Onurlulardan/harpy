package eventsv1

import "fmt"

func (catalog ConsumerCatalog) ResolveQueue(queueName string) (ConsumerPolicy, error) {
	var resolved *ConsumerPolicy
	for _, policy := range catalog.policies {
		if policy.QueueName != queueName {
			continue
		}
		if resolved != nil && (resolved.Consumer != policy.Consumer ||
			resolved.ConsumerVersion != policy.ConsumerVersion) {
			return ConsumerPolicy{}, fmt.Errorf("queue has multiple consumer owners")
		}
		copy := policy
		resolved = &copy
	}
	if resolved == nil {
		return ConsumerPolicy{}, fmt.Errorf("consumer queue is not registered")
	}
	return *resolved, nil
}
