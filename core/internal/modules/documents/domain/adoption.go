package domain

import "errors"

var ErrAdoptionTransition = errors.New("documents adoption transition is invalid")

type AdoptionState string

const (
	AdoptionInventory   AdoptionState = "inventory"
	AdoptionCopying     AdoptionState = "copying"
	AdoptionShadow      AdoptionState = "shadow"
	AdoptionReconciling AdoptionState = "reconciling"
	AdoptionQuiesced    AdoptionState = "quiesced"
	AdoptionAccepted    AdoptionState = "accepted"
	AdoptionBlocked     AdoptionState = "blocked"
	AdoptionReversed    AdoptionState = "reversed"
)

type AdoptionRun struct {
	State                                  AdoptionState
	SourceCount, TargetCount               uint64
	SourceBytes, TargetBytes               uint64
	SourceMerkleSHA256, TargetMerkleSHA256 [32]byte
	PendingObjectCount, PendingJobCount    uint64
	FinalLSN, AppliedLSN                   uint64
	AcceptanceSHA256                       [32]byte
}

func (run AdoptionRun) Advance(next AdoptionState) (AdoptionRun, error) {
	if !adoptionEdge(run.State, next) {
		return AdoptionRun{}, ErrAdoptionTransition
	}
	if next == AdoptionAccepted && !run.acceptable() {
		return AdoptionRun{}, ErrAdoptionTransition
	}
	run.State = next
	return run, nil
}

func (run AdoptionRun) acceptable() bool {
	return run.SourceCount == run.TargetCount && run.SourceBytes == run.TargetBytes &&
		run.SourceMerkleSHA256 != [32]byte{} && run.SourceMerkleSHA256 == run.TargetMerkleSHA256 &&
		run.PendingObjectCount == 0 && run.PendingJobCount == 0 && run.FinalLSN > 0 &&
		run.FinalLSN == run.AppliedLSN && run.AcceptanceSHA256 != [32]byte{}
}

func adoptionEdge(from, to AdoptionState) bool {
	switch from {
	case AdoptionInventory:
		return to == AdoptionCopying || to == AdoptionBlocked
	case AdoptionCopying:
		return to == AdoptionShadow || to == AdoptionBlocked
	case AdoptionShadow:
		return to == AdoptionReconciling || to == AdoptionBlocked
	case AdoptionReconciling:
		return to == AdoptionQuiesced || to == AdoptionBlocked
	case AdoptionQuiesced:
		return to == AdoptionAccepted || to == AdoptionReversed || to == AdoptionBlocked
	case AdoptionAccepted:
		return to == AdoptionReversed
	default:
		return false
	}
}
