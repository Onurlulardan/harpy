package deadletter

import "context"

type IngestStore interface {
	IngestEvent(context.Context, EventRoot) (IngestResult, error)
}
