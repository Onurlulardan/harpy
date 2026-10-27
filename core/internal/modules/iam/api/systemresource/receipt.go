package systemresource

type MutationReceipt struct {
	operationID       ID
	requestSHA256     [32]byte
	resultingRevision int64
	replayed          bool
	validated         bool
}

func NewMutationReceipt(
	operationID ID,
	requestSHA256 [32]byte,
	resultingRevision int64,
	replayed bool,
) (MutationReceipt, error) {
	receipt := MutationReceipt{
		operationID: operationID, requestSHA256: requestSHA256,
		resultingRevision: resultingRevision, replayed: replayed, validated: true,
	}
	if receipt.Validate() != nil {
		return MutationReceipt{}, ErrContract
	}
	return receipt, nil
}

func (receipt MutationReceipt) Validate() error {
	if !receipt.validated || !receipt.operationID.Valid() ||
		receipt.requestSHA256 == [32]byte{} || receipt.resultingRevision < 1 {
		return ErrContract
	}
	return nil
}

func (receipt MutationReceipt) OperationID() (ID, error) {
	if receipt.Validate() != nil {
		return ID{}, ErrContract
	}
	return receipt.operationID, nil
}

func (receipt MutationReceipt) RequestSHA256() ([32]byte, error) {
	if receipt.Validate() != nil {
		return [32]byte{}, ErrContract
	}
	return receipt.requestSHA256, nil
}

func (receipt MutationReceipt) Result() (int64, bool, error) {
	if receipt.Validate() != nil {
		return 0, false, ErrContract
	}
	return receipt.resultingRevision, receipt.replayed, nil
}
