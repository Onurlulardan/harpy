package contractretirement

import "context"

// RuntimePort separates read-only release admission from request hit evidence.
type RuntimePort interface {
	AdmitManifest(context.Context, Batch) (Decision, error)
	RecordHits(context.Context, Batch) (Decision, error)
}
