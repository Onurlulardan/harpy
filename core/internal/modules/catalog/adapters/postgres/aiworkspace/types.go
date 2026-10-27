package aiworkspace

import (
	"strings"
	"time"

	catalogai "github.com/Onurlulardan/harpy/core/internal/modules/catalog/api/aiworkspace"
	"github.com/google/uuid"
)

type menuItem struct {
	ItemID, PackageID uuid.UUID
	TargetID          *uuid.UUID
	Kind, Label       string
	Version           int64
	UpdatedAt         time.Time
	Score             float64
}

func (item menuItem) projection() (catalogai.Item, bool) {
	typeName, segment, slug, valid := item.identity()
	if !valid || strings.TrimSpace(item.Label) == "" || item.UpdatedAt.IsZero() {
		return catalogai.Item{}, false
	}
	return catalogai.Item{Path: segment + "/" + slug, Type: typeName, Slug: slug,
		Title: item.Label, UpdatedAt: item.UpdatedAt.UTC(), Score: item.Score,
		Requirements: ownerRequirements(item.ItemID)}, true
}

func (item menuItem) targetReference() (catalogai.TargetReference, bool) {
	if item.Kind == "GROUP" && item.ItemID != uuid.Nil {
		return catalogai.TargetReference{Kind: item.Kind, ID: item.ItemID}, true
	}
	if item.TargetID == nil || *item.TargetID == uuid.Nil {
		return catalogai.TargetReference{}, false
	}
	return catalogai.TargetReference{Kind: item.Kind, ID: *item.TargetID}, true
}

func (item menuItem) identity() (string, string, string, bool) {
	switch item.Kind {
	case "FORM":
		return targetIdentity("form", "forms", item.TargetID)
	case "WORKFLOW":
		return targetIdentity("workflow", "workflows", item.TargetID)
	case "DASHBOARD":
		return targetIdentity("dashboard", "dashboards", item.TargetID)
	case "ENDPOINT":
		return targetIdentity("endpoint", "endpoints", item.TargetID)
	case "GROUP":
		return "menu", "menus", item.ItemID.String(), item.ItemID != uuid.Nil
	default:
		return "", "", "", false
	}
}

func targetIdentity(typeName, segment string, id *uuid.UUID) (string, string, string, bool) {
	if id == nil || *id == uuid.Nil {
		return "", "", "", false
	}
	return typeName, segment, id.String(), true
}

func ownerRequirements(itemID uuid.UUID) []catalogai.Requirement {
	return []catalogai.Requirement{{Permission: "ai.workspace.read",
		Resource: "catalog/menu/" + itemID.String()}}
}

func normalizedKind(value string) string {
	switch strings.TrimSuffix(strings.ToLower(strings.TrimSpace(value)), "s") {
	case "form":
		return "FORM"
	case "workflow":
		return "WORKFLOW"
	case "dashboard":
		return "DASHBOARD"
	case "endpoint":
		return "ENDPOINT"
	case "menu":
		return "GROUP"
	default:
		return ""
	}
}
