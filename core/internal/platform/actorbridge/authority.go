package actorbridge

type Authority struct {
	iamGroupGeneration int64
	iamGeneration      int64
	targetGeneration   int64
}

func NewAuthority(iamGroupGeneration, iamGeneration, targetGeneration int64) (Authority, error) {
	authority := Authority{iamGroupGeneration: iamGroupGeneration,
		iamGeneration: iamGeneration, targetGeneration: targetGeneration}
	if authority.Validate() != nil {
		return Authority{}, NewFault(FaultInvalid)
	}
	return authority, nil
}

func (authority Authority) Validate() error {
	if authority.iamGroupGeneration <= 0 || authority.iamGeneration <= 0 ||
		authority.targetGeneration <= 0 {
		return NewFault(FaultInvalid)
	}
	return nil
}

func (authority Authority) IAMGroupGeneration() int64 { return authority.iamGroupGeneration }
func (authority Authority) IAMGeneration() int64      { return authority.iamGeneration }
func (authority Authority) TargetGeneration() int64   { return authority.targetGeneration }
