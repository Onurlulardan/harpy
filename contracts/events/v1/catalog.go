package eventsv1

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"

	canonicaljsonv1 "github.com/Onurlulardan/harpy/contracts/canonicaljson/v1"
)

type ConsumerCatalog struct {
	policies []ConsumerPolicy
}

func NewConsumerCatalog(policies []ConsumerPolicy) (ConsumerCatalog, error) {
	copy := append([]ConsumerPolicy(nil), policies...)
	sort.Slice(copy, func(left, right int) bool { return consumerPolicyLess(copy[left], copy[right]) })
	queueOwners := map[string]string{}
	ownerQueues := map[string]string{}
	ownerDomains := map[string]string{}
	ownerDeliveryLimits := map[string]int{}
	for index, policy := range copy {
		if err := policy.Validate(); err != nil {
			return ConsumerCatalog{}, fmt.Errorf("consumer policy %d: %w", index, err)
		}
		if index > 0 && sameConsumerPolicyKey(copy[index-1], policy) {
			return ConsumerCatalog{}, fmt.Errorf("duplicate consumer subscription policy")
		}
		owner := consumerQueueOwner(policy.Consumer, policy.ConsumerVersion)
		if registered, exists := queueOwners[policy.QueueName]; exists && registered != owner {
			return ConsumerCatalog{}, fmt.Errorf("queue cannot be shared across consumer versions")
		}
		if registered, exists := ownerQueues[owner]; exists && registered != policy.QueueName {
			return ConsumerCatalog{}, fmt.Errorf("consumer version must use one queue for all event types")
		}
		if registered, exists := ownerDomains[owner]; exists && registered != policy.OwnerDomain {
			return ConsumerCatalog{}, fmt.Errorf("consumer version must have one owner domain")
		}
		if registered, exists := ownerDeliveryLimits[owner]; exists && registered != policy.DeliveryLimit {
			return ConsumerCatalog{}, fmt.Errorf("consumer version must have one queue delivery limit")
		}
		queueOwners[policy.QueueName], ownerQueues[owner] = owner, policy.QueueName
		ownerDomains[owner] = policy.OwnerDomain
		ownerDeliveryLimits[owner] = policy.DeliveryLimit
	}
	return ConsumerCatalog{policies: copy}, nil
}

func (catalog ConsumerCatalog) ResolveIdentity(consumer string, consumerVersion int) (ConsumerPolicy, error) {
	for _, policy := range catalog.policies {
		if policy.Consumer == consumer && policy.ConsumerVersion == consumerVersion {
			return policy, nil
		}
	}
	return ConsumerPolicy{}, fmt.Errorf("consumer identity and version are not registered")
}

func (catalog ConsumerCatalog) Resolve(
	consumer string,
	consumerVersion int,
	eventType string,
	eventVersion int,
) (ConsumerPolicy, error) {
	wanted := ConsumerPolicy{
		Consumer: consumer, ConsumerVersion: consumerVersion,
		EventType: eventType, EventVersion: eventVersion,
	}
	index := sort.Search(len(catalog.policies), func(index int) bool {
		return !consumerPolicyLess(catalog.policies[index], wanted)
	})
	if index == len(catalog.policies) || !sameConsumerPolicyKey(catalog.policies[index], wanted) {
		return ConsumerPolicy{}, fmt.Errorf("consumer subscription is not registered")
	}
	return catalog.policies[index], nil
}

func (catalog ConsumerCatalog) Policies() []ConsumerPolicy {
	return append([]ConsumerPolicy(nil), catalog.policies...)
}

func (catalog ConsumerCatalog) CanonicalJSON() ([]byte, error) {
	document, err := json.Marshal(catalog.policies)
	if err != nil {
		return nil, err
	}
	return canonicaljsonv1.Canonicalize(document, MaximumDocumentBytes)
}

func (catalog ConsumerCatalog) SHA256() (string, error) {
	document, err := catalog.CanonicalJSON()
	if err != nil {
		return "", err
	}
	digest := sha256.Sum256(document)
	return hex.EncodeToString(digest[:]), nil
}

func consumerPolicyLess(left ConsumerPolicy, right ConsumerPolicy) bool {
	if left.Consumer != right.Consumer {
		return left.Consumer < right.Consumer
	}
	if left.ConsumerVersion != right.ConsumerVersion {
		return left.ConsumerVersion < right.ConsumerVersion
	}
	if left.EventType != right.EventType {
		return left.EventType < right.EventType
	}
	return left.EventVersion < right.EventVersion
}

func sameConsumerPolicyKey(left ConsumerPolicy, right ConsumerPolicy) bool {
	return left.Consumer == right.Consumer && left.ConsumerVersion == right.ConsumerVersion &&
		left.EventType == right.EventType && left.EventVersion == right.EventVersion
}

func consumerQueueOwner(consumer string, version int) string {
	return fmt.Sprintf("%s@%d", consumer, version)
}
