package v1query

import (
	"context"
	"fmt"
	"slices"

	"github.com/Onurlulardan/harpy/core/internal/platform/contractretirement"
)

type aliasEvidence struct {
	aliasID string
	key     contractretirement.Key
}

func (coordinator Coordinator) AdmitManifest(ctx context.Context) error {
	if ctx == nil || coordinator.Validate() != nil {
		return fmt.Errorf("V1 compatibility coordinator is invalid")
	}
	coordinator.admission.mutex.Lock()
	defer coordinator.admission.mutex.Unlock()
	coordinator.admission.admitted.Store(false)
	aliasIDs := coordinator.decoder.manifest.aliasIDs()
	if len(aliasIDs) == 0 {
		coordinator.admission.admitted.Store(true)
		return nil
	}
	batch, _, err := buildEvidenceBatch(
		coordinator.decoder.manifest, aliasIDs,
	)
	if err != nil {
		return err
	}
	decision, err := coordinator.evidence.AdmitManifest(ctx, batch)
	if err != nil || decision.Validate() != nil {
		if err != nil {
			return fmt.Errorf("admit V1 alias manifest: %w", err)
		}
		return fmt.Errorf("admit V1 alias manifest: invalid decision")
	}
	retired, err := decision.RetiredCount()
	if err != nil {
		return fmt.Errorf("admit V1 alias manifest: invalid decision")
	}
	if retired != 0 {
		return contractretirement.NewFault(contractretirement.FaultRetired)
	}
	coordinator.admission.admitted.Store(true)
	return nil
}

func (coordinator Coordinator) recordHits(
	ctx context.Context, pending pendingResult,
) (bool, error) {
	aliasIDs, err := pending.AliasIDs()
	if err != nil {
		return false, fmt.Errorf("V1 alias evidence is invalid")
	}
	if len(aliasIDs) == 0 {
		return false, nil
	}
	batch, evidence, err := buildEvidenceBatch(
		coordinator.decoder.manifest, aliasIDs,
	)
	if err != nil {
		return false, err
	}
	decision, err := coordinator.evidence.RecordHits(ctx, batch)
	if err != nil || decision.Validate() != nil {
		if err != nil {
			return false, fmt.Errorf("record V1 alias evidence: %w", err)
		}
		return false, fmt.Errorf("record V1 alias evidence: invalid decision")
	}
	collector := &issueCollector{}
	for _, used := range pending.aliasUses {
		index, found := slices.BinarySearchFunc(
			evidence, used.aliasID,
			func(item aliasEvidence, aliasID string) int {
				if item.aliasID < aliasID {
					return -1
				}
				if item.aliasID > aliasID {
					return 1
				}
				return 0
			},
		)
		if !found {
			return false, fmt.Errorf("V1 alias evidence is invalid")
		}
		retired, retiredErr := decision.Retires(evidence[index].key)
		if retiredErr != nil {
			return false, fmt.Errorf("V1 alias evidence is invalid")
		}
		if retired {
			collector.add(used.path, used.issue)
		}
	}
	if collector.hasIssues() {
		return false, collector.queryError()
	}
	return true, nil
}

func buildEvidenceBatch(
	manifest manifest, aliasIDs []string,
) (contractretirement.Batch, []aliasEvidence, error) {
	evidence := make([]aliasEvidence, len(aliasIDs))
	keys := make([]contractretirement.Key, len(aliasIDs))
	for index, aliasID := range aliasIDs {
		retirementKey, err := NewRetirementKey(manifest.id, aliasID)
		if err != nil {
			return contractretirement.Batch{}, nil, err
		}
		value, _ := retirementKey.Value()
		key, err := contractretirement.ParseKey(value)
		if err != nil {
			return contractretirement.Batch{}, nil, err
		}
		evidence[index], keys[index] = aliasEvidence{aliasID: aliasID, key: key}, key
	}
	batch, err := contractretirement.NewBatch(manifest.identity, keys)
	return batch, evidence, err
}
