package backupreceipt

import (
	"context"

	"github.com/google/uuid"
)

type Reader interface {
	Get(context.Context, uuid.UUID) (Receipt, error)
}

type Verifier struct{ reader Reader }

func NewVerifier(reader Reader) (*Verifier, error) {
	if reader == nil {
		return nil, ErrUnavailable
	}
	return &Verifier{reader: reader}, nil
}

func (verifier *Verifier) Verify(
	ctx context.Context,
	receiptID uuid.UUID,
	requirement Requirement,
) (Receipt, error) {
	if verifier == nil || verifier.reader == nil || ctx == nil || ctx.Err() != nil || receiptID == uuid.Nil {
		return Receipt{}, ErrUnavailable
	}
	receipt, err := verifier.reader.Get(ctx, receiptID)
	if err != nil {
		return Receipt{}, err
	}
	if receipt.ID != receiptID {
		return Receipt{}, ErrBinding
	}
	if err := receipt.ValidFor(requirement); err != nil {
		return Receipt{}, err
	}
	return receipt, nil
}
