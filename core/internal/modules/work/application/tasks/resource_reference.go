package tasks

import (
	"context"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/work/domain"
	"github.com/Onurlulardan/harpy/core/internal/modules/work/ports"
)

func (service *Service) registerTaskResource(
	ctx context.Context, scope ports.Scope, links []domain.ResourceLink,
) ([]domain.ResourceLink, error) {
	definition, found := definitionResourceLink(links)
	if !found {
		return links, nil
	}
	command, err := service.referenceCommand(definition)
	if err != nil {
		return nil, err
	}
	receipt, err := service.references.RegisterRead(ctx, scope, command)
	if err != nil || !receipt.Matches(command) {
		return nil, ports.ErrUnavailable
	}
	result := append([]domain.ResourceLink(nil), links...)
	for index := range result {
		result[index].ReferenceEpoch = receipt.ReferenceEpoch
		if !result[index].Valid() {
			return nil, ports.ErrUnavailable
		}
	}
	return result, nil
}

func (service *Service) replaceTaskResourceReference(
	ctx context.Context, scope ports.Scope,
	current, replacement []domain.ResourceLink,
) ([]domain.ResourceLink, error) {
	if definition, found := definitionResourceLink(current); found {
		command, err := service.referenceCommand(definition)
		if err != nil {
			return nil, err
		}
		receipt, releaseErr := service.references.ReleaseRead(ctx, scope, command)
		if releaseErr != nil || !receipt.Matches(command) ||
			receipt.ReferenceEpoch != definition.ReferenceEpoch {
			return nil, ports.ErrUnavailable
		}
	}
	return service.registerTaskResource(ctx, scope, replacement)
}

func (service *Service) referenceCommand(
	link domain.ResourceLink,
) (ports.ResourceReferenceCommand, error) {
	now := service.clock.Now().UTC().Truncate(time.Microsecond)
	operationID, err := service.ids.New(now)
	if err != nil {
		return ports.ResourceReferenceCommand{}, ports.ErrUnavailable
	}
	command := ports.ResourceReferenceCommand{Link: link,
		OperationID: operationID, OccurredAt: now}
	if !command.Link.ID.Valid() || !command.Link.WorkResourceID.Valid() ||
		!command.Link.TargetID.Valid() || !command.Link.TargetVersionID.Valid() ||
		command.Link.TargetOwner != domain.SourceForms ||
		command.Link.TargetKind != "FORM" || !command.OperationID.Valid() ||
		command.OccurredAt.IsZero() {
		return ports.ResourceReferenceCommand{}, ports.ErrInvalid
	}
	return command, nil
}

func definitionResourceLink(links []domain.ResourceLink) (domain.ResourceLink, bool) {
	for _, link := range links {
		if link.TargetOwner == domain.SourceForms && link.TargetKind == "FORM" {
			return link, true
		}
	}
	return domain.ResourceLink{}, false
}
