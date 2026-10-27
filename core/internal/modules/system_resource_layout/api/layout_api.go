package api

import (
	"context"
)

type LayoutAPI interface {
	Create(context.Context, CreateInput) (Revision, error)
	Get(context.Context, string) (Revision, error)
	List(context.Context, ResourceKind) ([]Revision, error)
	Update(context.Context, UpdateInput) (Revision, error)
	Delete(context.Context, DeleteInput) error
}
