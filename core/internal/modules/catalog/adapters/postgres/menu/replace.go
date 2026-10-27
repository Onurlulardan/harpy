package menu

import (
	"context"
	_ "embed"
	"math"

	menuapp "github.com/Onurlulardan/harpy/core/internal/modules/catalog/application/menu"
	menudomain "github.com/Onurlulardan/harpy/core/internal/modules/catalog/domain/menu"
	"github.com/Onurlulardan/harpy/core/internal/modules/catalog/domain/packagecontract"
	"github.com/google/uuid"
)

//go:embed sql/tenant/lock_definition.sql
var lockDefinitionSQL string

//go:embed sql/tenant/create_version.sql
var createVersionSQL string

//go:embed sql/tenant/insert_item.sql
var insertItemSQL string

//go:embed sql/tenant/activate_version.sql
var activateVersionSQL string

//go:embed sql/tenant/reveal_item_override.sql
var revealItemOverrideSQL string

//go:embed sql/tenant/insert_hidden_override.sql
var insertHiddenOverrideSQL string

func (storage *store) Replace(
	ctx context.Context, mutation menuapp.ReplaceMutation,
) (menuapp.Source, error) {
	if !validReplace(mutation) {
		return menuapp.Source{}, menuapp.ErrInvalid
	}
	if !storage.usable(ctx, true) {
		return menuapp.Source{}, unavailable()
	}
	var revision int64
	var active uuid.NullUUID
	if err := storage.queryRow(ctx, lockDefinitionSQL, menuDefinitionID).Scan(
		&revision, &active); err != nil || revision < 1 {
		return menuapp.Source{}, unavailable()
	}
	if uint64(revision) != mutation.ExpectedRevision {
		return menuapp.Source{}, menuapp.ErrConflict
	}
	versionNumber := revision + 1
	if _, err := storage.execute(ctx, createVersionSQL, mutation.VersionID.String(),
		menuDefinitionID, versionNumber, mutation.ContentSHA256, mutation.OccurredAt); err != nil {
		return menuapp.Source{}, unavailable()
	}
	for _, item := range mutation.Items {
		if err := storage.insertReplaceItem(ctx, mutation.VersionID.String(), item); err != nil {
			return menuapp.Source{}, err
		}
	}
	overrides, err := storage.replaceOverrides(ctx, mutation.Items)
	if err != nil {
		return menuapp.Source{}, err
	}
	rows, err := storage.execute(ctx, activateVersionSQL, menuDefinitionID,
		mutation.VersionID.String(), versionNumber, mutation.OccurredAt, revision)
	if err != nil || rows != 1 {
		return menuapp.Source{}, unavailable()
	}
	return replaceSource(mutation, uint64(versionNumber), overrides)
}

func (storage *store) insertReplaceItem(
	ctx context.Context, versionID string, item menudomain.Item,
) error {
	var parentID, targetID, externalURL any
	if item.ParentID != nil {
		parentID = item.ParentID.String()
	}
	if item.TargetID != nil {
		targetID = item.TargetID.String()
	}
	if item.URL != "" {
		externalURL = item.URL
	}
	if _, err := storage.execute(ctx, insertItemSQL, versionID, item.ID.String(),
		parentID, item.PackageID.String(), string(item.Kind), targetID, externalURL,
		item.Label, item.Order); err != nil {
		return unavailable()
	}
	return nil
}

func (storage *store) replaceOverrides(
	ctx context.Context, items []menudomain.Item,
) ([]menudomain.Override, error) {
	overrides := make([]menudomain.Override, 0)
	for _, item := range items {
		if item.Visible {
			if _, err := storage.execute(ctx, revealItemOverrideSQL, item.ID.String()); err != nil {
				return nil, unavailable()
			}
			continue
		}
		if _, err := storage.execute(ctx, insertHiddenOverrideSQL, item.ID.String()); err != nil {
			return nil, unavailable()
		}
		visible := false
		overrides = append(overrides, menudomain.Override{ItemID: item.ID, Visible: &visible})
	}
	return overrides, nil
}

func validReplace(value menuapp.ReplaceMutation) bool {
	if value.VersionID == uuid.Nil || value.VersionID.Version() != 7 ||
		value.ExpectedRevision < 1 || value.ExpectedRevision > math.MaxInt64 ||
		len(value.Items) > 2048 || !digestPattern.MatchString(value.ContentSHA256) ||
		!canonicalTime(value.OccurredAt) {
		return false
	}
	seen := map[string]bool{}
	for _, item := range value.Items {
		if !item.ID.Valid() || !item.PackageID.Valid() || seen[item.ID.String()] {
			return false
		}
		seen[item.ID.String()] = true
	}
	return validParents(value.Items, seen)
}

func replaceSource(
	mutation menuapp.ReplaceMutation, number uint64, overrides []menudomain.Override,
) (menuapp.Source, error) {
	id, err := packagecontract.ParseStableID(mutation.VersionID.String())
	if err != nil {
		return menuapp.Source{}, unavailable()
	}
	items := append([]menudomain.Item(nil), mutation.Items...)
	for index := range items {
		items[index].Visible = true
	}
	return menuapp.Source{
		Version:   menudomain.Version{ID: id, Number: number, Revision: number, Items: items},
		Overrides: overrides, ActivePackages: activateTenantAuthoredPackages(),
	}, nil
}

func activateTenantAuthoredPackages() map[string]bool {
	return map[string]bool{menuapp.TenantAuthoredPackageID: true}
}
