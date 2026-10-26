package work

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/work/application/tasks"
	"github.com/Onurlulardan/harpy/core/internal/modules/work/ports"
)

type LegacyTaskResourceRequest struct {
	FormSlug   string
	FormDataID string
}

type LegacyTaskResourceDescription struct {
	FormSlug   string
	FormDataID string
}

type LegacyTaskResourceResolver interface {
	Resolve(context.Context, ports.Scope, LegacyTaskResourceRequest) (
		*tasks.ResourceLinkInput, error)
	Describe(context.Context, ports.Scope, *tasks.ResourceLinkInput) (
		LegacyTaskResourceDescription, error)
}

type closedTaskReferences struct{}

func NewClosedTaskReferences() (ports.ResourceReferenceRegistry, error) {
	return closedTaskReferences{}, nil
}

func (closedTaskReferences) RegisterRead(context.Context, ports.Scope,
	ports.ResourceReferenceCommand) (ports.ResourceReferenceReceipt, error) {
	return ports.ResourceReferenceReceipt{}, ports.ErrUnavailable
}

func (closedTaskReferences) ReleaseRead(context.Context, ports.Scope,
	ports.ResourceReferenceCommand) (ports.ResourceReferenceReceipt, error) {
	return ports.ResourceReferenceReceipt{}, ports.ErrUnavailable
}

type closedTaskResources struct{}

func NewClosedTaskResources() (LegacyTaskResourceResolver, error) {
	return closedTaskResources{}, nil
}

func (closedTaskResources) Resolve(context.Context, ports.Scope,
	LegacyTaskResourceRequest) (*tasks.ResourceLinkInput, error) {
	return nil, ports.ErrUnavailable
}

func (closedTaskResources) Describe(context.Context, ports.Scope,
	*tasks.ResourceLinkInput) (LegacyTaskResourceDescription, error) {
	return LegacyTaskResourceDescription{}, ports.ErrUnavailable
}
