package eventsv1

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"

	canonicaljsonv1 "github.com/Onurlulardan/harpy/contracts/canonicaljson/v1"
)

type ProducerCatalog struct {
	policies []ProducerPolicy
}

func NewProducerCatalog(policies []ProducerPolicy) (ProducerCatalog, error) {
	copy := append([]ProducerPolicy(nil), policies...)
	sort.Slice(copy, func(left, right int) bool {
		if copy[left].EventType == copy[right].EventType {
			return copy[left].EventVersion < copy[right].EventVersion
		}
		return copy[left].EventType < copy[right].EventType
	})
	for index, policy := range copy {
		if err := policy.Validate(); err != nil {
			return ProducerCatalog{}, fmt.Errorf("producer policy %d: %w", index, err)
		}
		if index > 0 && copy[index-1].EventType == policy.EventType &&
			copy[index-1].EventVersion == policy.EventVersion {
			return ProducerCatalog{}, fmt.Errorf("duplicate producer event version policy")
		}
	}
	return ProducerCatalog{policies: copy}, nil
}

func (catalog ProducerCatalog) Resolve(eventType string, version int) (ProducerPolicy, error) {
	index := sort.Search(len(catalog.policies), func(index int) bool {
		policy := catalog.policies[index]
		return policy.EventType > eventType || policy.EventType == eventType && policy.EventVersion >= version
	})
	if index == len(catalog.policies) || catalog.policies[index].EventType != eventType ||
		catalog.policies[index].EventVersion != version {
		return ProducerPolicy{}, fmt.Errorf("producer event type and version are not registered")
	}
	return catalog.policies[index], nil
}

func (catalog ProducerCatalog) Policies() []ProducerPolicy {
	return append([]ProducerPolicy(nil), catalog.policies...)
}

func (catalog ProducerCatalog) CanonicalJSON() ([]byte, error) {
	document, err := json.Marshal(catalog.policies)
	if err != nil {
		return nil, err
	}
	return canonicaljsonv1.Canonicalize(document, MaximumDocumentBytes)
}

func (catalog ProducerCatalog) SHA256() (string, error) {
	document, err := catalog.CanonicalJSON()
	if err != nil {
		return "", err
	}
	digest := sha256.Sum256(document)
	return hex.EncodeToString(digest[:]), nil
}
