package v1adoption

import (
	"context"
	"encoding/json"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/work/domain"
	"github.com/Onurlulardan/harpy/core/internal/modules/work/ports"
)

type Source interface {
	LoadLegacy(context.Context, ports.Scope) ([]SourceRecord, error)
}

type Target interface {
	PersistPlan(context.Context, ports.Scope, Plan) (Receipt, error)
	ComparePlan(context.Context, ports.Scope, Plan) (Evidence, error)
	LoadEvidence(context.Context, ports.Scope, domain.ID) (Evidence, error)
	RevertPlan(context.Context, ports.Scope, ReverseRequest, Evidence) (Receipt, error)
	SealLegacy(context.Context, ports.Scope, RetireRequest, Evidence) (Receipt, error)
}

type FormReference struct {
	ID, VersionID      domain.ID
	PermissionRevision int64
}

type FormResolver interface {
	ResolveForm(context.Context, ports.Scope, string, *domain.ID, json.RawMessage) (FormReference, error)
	RegisterRead(context.Context, ports.Scope, ports.ResourceReferenceCommand) (
		ports.ResourceReferenceReceipt, error)
}

type WorkflowProjection struct {
	TaskID, VersionID            domain.ID
	Revision, PermissionRevision int64
	Summary                      string
	DueAt                        *time.Time
	VisibleTo                    []domain.Principal
}

type WorkflowResolver interface {
	ResolveTask(context.Context, ports.Scope, domain.ID) (WorkflowProjection, error)
}

type Clock interface{ Now() time.Time }

type Dependencies struct {
	Source      Source
	Target      Target
	Authorizer  ports.Authorizer
	Forms       FormResolver
	Workflow    WorkflowResolver
	Principals  ports.PrincipalResolver
	TZDB        ports.TZDBRegistry
	TZDBVersion string
	Clock       Clock
}

type Service struct{ dependencies Dependencies }

func NewService(dependencies Dependencies) (*Service, error) {
	if dependencies.Source == nil || dependencies.Target == nil ||
		dependencies.Authorizer == nil || dependencies.Forms == nil ||
		dependencies.Workflow == nil || dependencies.Principals == nil || dependencies.TZDB == nil ||
		dependencies.TZDBVersion == "" || dependencies.Clock == nil {
		return nil, ports.ErrInvalid
	}
	return &Service{dependencies: dependencies}, nil
}
