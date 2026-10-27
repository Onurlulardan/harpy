package systemresource

type MutationResult struct {
	read      ReadResult
	receipt   MutationReceipt
	validated bool
}

func NewMutationResult(
	read ReadResult,
	receipt MutationReceipt,
) (MutationResult, error) {
	result := MutationResult{read: read, receipt: receipt, validated: true}
	if result.Validate() != nil {
		return MutationResult{}, ErrContract
	}
	return result, nil
}

func (result MutationResult) Validate() error {
	if !result.validated || result.read.Validate() != nil ||
		result.receipt.Validate() != nil {
		return ErrContract
	}
	record, _ := result.read.Record()
	_, revision, recordErr := record.Identity()
	receiptRevision, replayed, receiptErr := result.receipt.Result()
	if recordErr != nil || receiptErr != nil || revision < receiptRevision ||
		(!replayed && revision != receiptRevision) {
		return ErrContract
	}
	return nil
}

func (result MutationResult) Read() (ReadResult, error) {
	if result.Validate() != nil {
		return ReadResult{}, ErrContract
	}
	return result.read, nil
}

func (result MutationResult) Receipt() (MutationReceipt, error) {
	if result.Validate() != nil {
		return MutationReceipt{}, ErrContract
	}
	return result.receipt, nil
}
