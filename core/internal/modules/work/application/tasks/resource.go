package tasks

import (
	"context"
	"strings"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/work/domain"
	"github.com/Onurlulardan/harpy/core/internal/modules/work/ports"
)

func assigneeFingerprint(values []domain.Principal, revision int64) string {
	parts := []string{ports.VersionString(revision)}
	for _, principal := range values {
		parts = append(parts, string(principal.Kind), principal.ID.String())
	}
	return strings.Join(parts, ":")
}

func resourceMatchesKind(kind domain.TaskKind, link *ResourceLinkInput) bool {
	if kind == domain.TaskCustom {
		return link == nil
	}
	if link == nil || link.Owner != domain.SourceForms ||
		len(link.Targets) < 1 || len(link.Targets) > 2 {
		return false
	}
	definition := link.Targets[0]
	if !validResourceTarget(definition) || definition.Kind != "FORM" {
		return false
	}
	if len(link.Targets) == 1 {
		return true
	}
	record := link.Targets[1]
	return validResourceTarget(record) && record.Kind == "FORM_RECORD" &&
		record.VersionID == definition.VersionID &&
		record.ID != definition.ID
}

func validResourceTarget(target ResourceTargetInput) bool {
	return target.ID.Valid() && target.VersionID.Valid() &&
		strings.TrimSpace(target.Kind) != "" && len(target.Kind) <= 64
}

func resourceInputFromLinks(
	taskID domain.ID, links []domain.ResourceLink,
) (*ResourceLinkInput, error) {
	if len(links) == 0 {
		return nil, nil
	}
	input := &ResourceLinkInput{Owner: links[0].TargetOwner,
		Targets: make([]ResourceTargetInput, len(links))}
	for index, link := range links {
		if !link.Valid() || link.WorkResourceID != taskID ||
			link.TargetOwner != input.Owner {
			return nil, ports.ErrUnavailable
		}
		input.Targets[index] = ResourceTargetInput{Kind: link.TargetKind,
			ID: link.TargetID, VersionID: link.TargetVersionID}
	}
	return input, nil
}

// ResourceInputFromLinks restores the exact canonical owner targets held by a
// Work task. HTTP compatibility adapters use it only while the task row and
// its resource links are locked in the same tenant transaction.
func ResourceInputFromLinks(
	taskID domain.ID, links []domain.ResourceLink,
) (*ResourceLinkInput, error) {
	return resourceInputFromLinks(taskID, links)
}

func validStoredTaskResource(
	task domain.Task, links []domain.ResourceLink,
) bool {
	input, err := resourceInputFromLinks(task.ID, links)
	return err == nil && resourceMatchesKind(task.Kind, input)
}

func validResourceReplacement(
	replace bool, kind domain.TaskKind, link *ResourceLinkInput,
) bool {
	if !replace {
		return kind == "" && link == nil
	}
	return kind.Valid() && resourceMatchesKind(kind, link)
}

func (service *Service) checkTaskResource(ctx context.Context, scope ports.Scope,
	link *ResourceLinkInput,
) ([]int64, error) {
	if link == nil {
		return nil, nil
	}
	revisions := make([]int64, len(link.Targets))
	for index, target := range link.Targets {
		allowed, revision, err := service.resources.AuthorizeRead(ctx, scope,
			ports.ResourceAuthorityRequest{Owner: link.Owner, Kind: target.Kind,
				ID: target.ID, VersionID: target.VersionID})
		if err != nil || !allowed || revision < 1 {
			return nil, ports.ErrForbidden
		}
		revisions[index] = revision
	}
	return revisions, nil
}

func resourceFingerprint(link *ResourceLinkInput) string {
	if link == nil {
		return ""
	}
	parts := []string{string(link.Owner)}
	for _, target := range link.Targets {
		parts = append(parts, target.Kind, target.ID.String(), target.VersionID.String())
	}
	return strings.Join(parts, ":")
}

func resourceAuthorityFingerprint(
	link *ResourceLinkInput, revisions []int64,
) string {
	if link == nil {
		return ""
	}
	parts := []string{resourceFingerprint(link)}
	for _, revision := range revisions {
		parts = append(parts, ports.VersionString(revision))
	}
	return strings.Join(parts, ":")
}

func resourceReplacementFingerprint(
	replace bool, kind domain.TaskKind, link *ResourceLinkInput, revisions []int64,
) string {
	if !replace {
		return "resource:preserve"
	}
	return "resource:replace:" + string(kind) + ":" +
		resourceAuthorityFingerprint(link, revisions)
}

func (service *Service) newResourceLinks(
	taskID domain.ID, input *ResourceLinkInput, permissionRevisions []int64,
) ([]domain.ResourceLink, error) {
	if input == nil {
		return nil, nil
	}
	if len(input.Targets) != len(permissionRevisions) {
		return nil, ports.ErrInvalid
	}
	links := make([]domain.ResourceLink, len(input.Targets))
	for index, target := range input.Targets {
		linkID, err := service.ids.New(service.clock.Now().UTC())
		if err != nil {
			return nil, ports.ErrUnavailable
		}
		links[index] = domain.ResourceLink{ID: linkID, WorkResourceID: taskID,
			TargetID: target.ID, TargetVersionID: target.VersionID,
			TargetOwner: input.Owner, TargetKind: target.Kind,
			PermissionRevision: permissionRevisions[index],
			RowVersion:         1}
		if !links[index].ID.Valid() || !links[index].WorkResourceID.Valid() ||
			!links[index].TargetID.Valid() || !links[index].TargetVersionID.Valid() ||
			!links[index].TargetOwner.Valid() || links[index].TargetKind == "" ||
			links[index].PermissionRevision < 1 {
			return nil, ports.ErrInvalid
		}
	}
	return links, nil
}

func timeString(value *time.Time) string {
	if value == nil {
		return ""
	}
	return value.UTC().Format(time.RFC3339Nano)
}
