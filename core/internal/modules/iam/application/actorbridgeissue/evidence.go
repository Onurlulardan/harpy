package actorbridgeissue

type Evidence struct {
	iamGroupGeneration int64
	iamGeneration      int64
	targetGeneration   int64
}

func NewEvidence(
	iamGroupGeneration int64,
	iamGeneration int64,
	targetGeneration int64,
) (Evidence, error) {
	if iamGroupGeneration < 1 || iamGeneration < 1 || targetGeneration < 1 {
		return Evidence{}, ErrUnavailable
	}
	return Evidence{iamGroupGeneration: iamGroupGeneration,
		iamGeneration: iamGeneration, targetGeneration: targetGeneration}, nil
}

func (evidence Evidence) IAMGroupGeneration() int64 { return evidence.iamGroupGeneration }
func (evidence Evidence) IAMGeneration() int64      { return evidence.iamGeneration }
func (evidence Evidence) TargetGeneration() int64   { return evidence.targetGeneration }

func (evidence Evidence) Validate() error {
	if evidence.iamGroupGeneration < 1 || evidence.iamGeneration < 1 ||
		evidence.targetGeneration < 1 {
		return ErrUnavailable
	}
	return nil
}
