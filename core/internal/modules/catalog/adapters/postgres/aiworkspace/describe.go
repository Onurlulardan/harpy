package aiworkspace

import (
	"context"
	"strings"

	catalogai "github.com/Onurlulardan/harpy/core/internal/modules/catalog/api/aiworkspace"
	"github.com/google/uuid"
)

func (owner *Owner) Describe(ctx context.Context, path string,
) (catalogai.Detail, error) {
	kind, identity, err := parsePath(path)
	if !owner.allowed(ctx) {
		return catalogai.Detail{}, catalogai.ErrUnavailable
	}
	if err != nil {
		return catalogai.Detail{}, catalogai.ErrInvalid
	}
	item, err := owner.readOne(ctx, describeSQL, identity, kind)
	if err != nil {
		return catalogai.Detail{}, err
	}
	projection, valid := item.projection()
	target, targetValid := item.targetReference()
	if !valid || !targetValid || projection.Path != path {
		return catalogai.Detail{}, catalogai.ErrUnavailable
	}
	return catalogai.Detail{Path: projection.Path, Type: projection.Type,
		Slug: projection.Slug, Title: projection.Title, UpdatedAt: projection.UpdatedAt,
		Target: target, Requirements: projection.Requirements}, nil
}

func (owner *Owner) readOne(ctx context.Context, statement string,
	arguments ...any,
) (menuItem, error) {
	item, err := scanMenuItem(owner.queryRow(ctx, statement, arguments...))
	if err != nil {
		if owner.notFound(err) {
			return menuItem{}, catalogai.ErrNotFound
		}
		return menuItem{}, catalogai.ErrUnavailable
	}
	return item, nil
}

func parsePath(path string) (string, uuid.UUID, error) {
	parts := strings.Split(path, "/")
	if len(parts) != 2 || path != strings.TrimSpace(path) {
		return "", uuid.Nil, catalogai.ErrInvalid
	}
	kind := normalizedKind(parts[0])
	identity, err := uuid.Parse(parts[1])
	if kind == "" || err != nil || identity == uuid.Nil || identity.String() != parts[1] {
		return "", uuid.Nil, catalogai.ErrInvalid
	}
	return kind, identity, nil
}
