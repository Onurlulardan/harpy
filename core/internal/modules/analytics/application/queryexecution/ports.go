package queryexecution

import (
	"context"
)

type Runner interface {
	Run(context.Context, Item) ([]byte, error)
}
