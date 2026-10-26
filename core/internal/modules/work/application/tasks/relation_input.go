package tasks

import (
	"strings"

	"github.com/Onurlulardan/harpy/core/internal/modules/work/domain"
	"github.com/Onurlulardan/harpy/core/internal/modules/work/ports"
)

type ChecklistValue struct {
	ID        *domain.ID
	Text      string
	Completed bool
}

type RelationInput struct {
	ReplaceAssignees bool
	Assignees        []domain.Principal
	ReplaceLabels    bool
	Labels           []string
	ReplaceChecklist bool
	Checklist        []ChecklistValue
}

func normalizeRelations(input *RelationInput) error {
	if input == nil {
		return ports.ErrInvalid
	}
	if len(input.Assignees) > 100 || len(input.Labels) > 100 || len(input.Checklist) > 1000 {
		return ports.ErrInvalid
	}
	labels := map[string]bool{}
	for index := range input.Labels {
		input.Labels[index] = strings.TrimSpace(input.Labels[index])
		if input.Labels[index] == "" || len(input.Labels[index]) > 100 || labels[input.Labels[index]] {
			return ports.ErrInvalid
		}
		labels[input.Labels[index]] = true
	}
	ids := map[domain.ID]bool{}
	for index := range input.Checklist {
		input.Checklist[index].Text = strings.TrimSpace(input.Checklist[index].Text)
		if input.Checklist[index].Text == "" || len(input.Checklist[index].Text) > 1000 {
			return ports.ErrInvalid
		}
		if input.Checklist[index].ID != nil {
			if !input.Checklist[index].ID.Valid() || ids[*input.Checklist[index].ID] {
				return ports.ErrInvalid
			}
			ids[*input.Checklist[index].ID] = true
		}
	}
	return nil
}

func relationFingerprint(input RelationInput) []string {
	parts := []string{ports.VersionString(boolInt(input.ReplaceAssignees)),
		ports.VersionString(boolInt(input.ReplaceLabels)),
		ports.VersionString(boolInt(input.ReplaceChecklist))}
	for _, label := range input.Labels {
		parts = append(parts, "label", label)
	}
	for _, item := range input.Checklist {
		parts = append(parts, "check", ports.IDString(item.ID), item.Text,
			ports.VersionString(boolInt(item.Completed)))
	}
	return parts
}

func boolInt(value bool) int64 {
	if value {
		return 1
	}
	return 0
}
