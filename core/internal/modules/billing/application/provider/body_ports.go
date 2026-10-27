package providerapp

import "context"

type BodyProtector interface {
	Protect(context.Context, []byte, []byte) (ProtectedBody, error)
}

type ReconciliationBodyOpener interface {
	OpenForReconciliation(context.Context, []byte, ProtectedBody) ([]byte, error)
}
