package adoption

import (
	"strings"

	adoptionapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/adoption"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/domain"
)

func (storage *Store) scanRun(row Row) (adoptionapp.Run, bool, error) {
	var result adoptionapp.Run
	var id, state string
	var generation int64
	var manifest, sourceRoot, targetRoot, acceptance []byte
	var sourceCount, targetCount, sourceBytes, targetBytes int64
	var pendingObjects, pendingJobs int64
	var finalLSN, appliedLSN dbText
	var replay bool
	err := row.Scan(&id, &generation, &manifest, &state, &sourceCount,
		&targetCount, &sourceBytes, &targetBytes, &sourceRoot, &targetRoot,
		&pendingObjects, &pendingJobs, &finalLSN, &appliedLSN, &acceptance,
		&result.CreatedAt, &result.UpdatedAt, &replay)
	if err != nil || generation < 1 || sourceCount < 0 || targetCount < 0 ||
		sourceBytes < 0 || targetBytes < 0 || pendingObjects < 0 || pendingJobs < 0 ||
		len(manifest) != 32 {
		return adoptionapp.Run{}, false, adoptionapp.ErrConflict
	}
	result.RunID, err = domain.ParseID(id)
	if err != nil {
		return adoptionapp.Run{}, false, adoptionapp.ErrUnavailable
	}
	result.TenantID, result.CutoverGeneration = storage.tenantID, uint64(generation)
	copy(result.ManifestSHA256[:], manifest)
	result.State, err = adoptionState(state)
	if err != nil {
		return adoptionapp.Run{}, false, err
	}
	result.SourceCount, result.TargetCount = uint64(sourceCount), uint64(targetCount)
	result.SourceBytes, result.TargetBytes = uint64(sourceBytes), uint64(targetBytes)
	result.PendingObjectCount, result.PendingJobCount = uint64(pendingObjects), uint64(pendingJobs)
	copyOptionalDigest(&result.SourceMerkleSHA256, sourceRoot)
	copyOptionalDigest(&result.TargetMerkleSHA256, targetRoot)
	copyOptionalDigest(&result.AcceptanceSHA256, acceptance)
	result.FinalLSN, result.AppliedLSN = finalLSN.Text, appliedLSN.Text
	return result, replay, nil
}

func adoptionState(value string) (domain.AdoptionState, error) {
	state := domain.AdoptionState(strings.ToLower(value))
	switch state {
	case domain.AdoptionInventory, domain.AdoptionCopying, domain.AdoptionShadow,
		domain.AdoptionReconciling, domain.AdoptionQuiesced, domain.AdoptionAccepted,
		domain.AdoptionBlocked, domain.AdoptionReversed:
		return state, nil
	default:
		return "", adoptionapp.ErrUnavailable
	}
}

func copyOptionalDigest(target *[32]byte, source []byte) {
	if len(source) == len(target) {
		copy(target[:], source)
	}
}
