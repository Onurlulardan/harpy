package customerworker

import "context"

type ClaimStore interface {
	ClaimNext(context.Context, ClaimRequest) (Claim, error)
}

type AuthorityAcknowledger interface {
	AcknowledgePrepared(context.Context, string) (AuthorityReceipt, error)
}

type AuthorityRecorder interface {
	RecordAuthority(context.Context, Claim, AuthorityReceipt) (int64, error)
}

type Finalizer interface {
	Finalize(context.Context, Claim, int64) error
}
