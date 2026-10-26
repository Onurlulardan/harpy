package jobsv1

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
)

type Catalog struct {
	policies []KindPolicy
}

func NewCatalog(policies []KindPolicy) (Catalog, error) {
	copy := append([]KindPolicy(nil), policies...)
	sort.Slice(copy, func(left, right int) bool {
		if copy[left].Kind == copy[right].Kind {
			return copy[left].PayloadRevision < copy[right].PayloadRevision
		}
		return copy[left].Kind < copy[right].Kind
	})
	for index, policy := range copy {
		if err := policy.Validate(); err != nil {
			return Catalog{}, fmt.Errorf("job policy %d: %w", index, err)
		}
		if index > 0 && copy[index-1].Kind == policy.Kind &&
			copy[index-1].PayloadRevision == policy.PayloadRevision {
			return Catalog{}, fmt.Errorf("duplicate job kind policy")
		}
	}
	return Catalog{policies: copy}, nil
}

func (catalog Catalog) Resolve(kind string, revision int) (KindPolicy, error) {
	index := sort.Search(len(catalog.policies), func(index int) bool {
		policy := catalog.policies[index]
		return policy.Kind > kind || policy.Kind == kind && policy.PayloadRevision >= revision
	})
	if index == len(catalog.policies) || catalog.policies[index].Kind != kind ||
		catalog.policies[index].PayloadRevision != revision {
		return KindPolicy{}, fmt.Errorf("job kind and payload revision are not registered")
	}
	return catalog.policies[index], nil
}

func (catalog Catalog) Policies() []KindPolicy {
	return append([]KindPolicy(nil), catalog.policies...)
}

func (catalog Catalog) CanonicalJSON() ([]byte, error) {
	return json.Marshal(catalog.policies)
}

func (catalog Catalog) SHA256() (string, error) {
	document, err := catalog.CanonicalJSON()
	if err != nil {
		return "", err
	}
	digest := sha256.Sum256(document)
	return hex.EncodeToString(digest[:]), nil
}
