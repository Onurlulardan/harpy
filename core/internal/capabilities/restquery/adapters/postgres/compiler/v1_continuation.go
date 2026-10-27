package compiler

import "crypto/sha256"

func v1PredecessorBinding(
	base statementBinding, predecessorIndex int64,
) (statementBinding, error) {
	if base.validate() != nil || predecessorIndex < 0 {
		return statementBinding{}, ErrContract
	}
	writer := newSealWriter("harpy.restquery.postgres.v1-predecessor.v1")
	writer.writeBytes(base.continuation[:])
	writer.writeInt(predecessorIndex)
	continuation, err := writer.sum()
	if err != nil {
		return statementBinding{}, ErrContract
	}
	result := base
	result.continuation = continuation
	if result.validate() != nil {
		return statementBinding{}, ErrContract
	}
	return result, nil
}

func v1PageBinding(
	base statementBinding,
	predecessorSeal [sha256.Size]byte,
	tupleSeal [sha256.Size]byte,
) (statementBinding, error) {
	if base.validate() != nil || predecessorSeal == [sha256.Size]byte{} ||
		tupleSeal == [sha256.Size]byte{} {
		return statementBinding{}, ErrContract
	}
	writer := newSealWriter("harpy.restquery.postgres.v1-keyset-page.v1")
	writer.writeBytes(base.continuation[:])
	writer.writeBytes(predecessorSeal[:])
	writer.writeBytes(tupleSeal[:])
	continuation, err := writer.sum()
	if err != nil {
		return statementBinding{}, ErrContract
	}
	result := base
	result.continuation = continuation
	if result.validate() != nil {
		return statementBinding{}, ErrContract
	}
	return result, nil
}
