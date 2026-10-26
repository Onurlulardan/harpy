package application

import "context"

type RetentionStore interface {
	Prune(context.Context, RetentionPrune) (int, error)
}
