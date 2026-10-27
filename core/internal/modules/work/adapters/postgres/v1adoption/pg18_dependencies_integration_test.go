//go:build integration

package v1adoption

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	adoption "github.com/Onurlulardan/harpy/core/internal/modules/work/application/v1adoption"
	"github.com/Onurlulardan/harpy/core/internal/modules/work/domain"
	"github.com/Onurlulardan/harpy/core/internal/modules/work/ports"
)

type workAdoptionScope struct{ tenant, actor domain.ID }

func (scope workAdoptionScope) TenantID() domain.ID { return scope.tenant }
func (scope workAdoptionScope) ActorID() domain.ID  { return scope.actor }
func (workAdoptionScope) OperationID() string       { return "work-pg18-adoption" }
func (workAdoptionScope) SnapshotQuery() bool       { return false }

type workAdoptionAuthorizer struct{}

func (workAdoptionAuthorizer) Authorize(
	context.Context, ports.Scope, string, domain.ID,
) error {
	return nil
}

type workAdoptionForms struct{}

func (workAdoptionForms) ResolveForm(context.Context, ports.Scope, string,
	*domain.ID, json.RawMessage,
) (adoption.FormReference, error) {
	return adoption.FormReference{ID: workDomainID(31), VersionID: workDomainID(32),
		PermissionRevision: 1}, nil
}

func (workAdoptionForms) RegisterRead(_ context.Context, _ ports.Scope,
	command ports.ResourceReferenceCommand,
) (ports.ResourceReferenceReceipt, error) {
	return ports.ResourceReferenceReceipt{ReferenceID: command.Link.ID,
		OperationID: command.OperationID, TargetID: command.Link.TargetID,
		TargetVersionID: command.Link.TargetVersionID, ReferenceEpoch: 1}, nil
}

type workAdoptionWorkflow struct{}

func (workAdoptionWorkflow) ResolveTask(context.Context, ports.Scope,
	domain.ID,
) (adoption.WorkflowProjection, error) {
	return adoption.WorkflowProjection{TaskID: workDomainID(33), VersionID: workDomainID(34),
		Revision: 1, PermissionRevision: 1, Summary: "Work PG18",
		VisibleTo: []domain.Principal{{Kind: domain.PrincipalUser, ID: workDomainID(1)}}}, nil
}

type workAdoptionPrincipals struct{}

func (workAdoptionPrincipals) Resolve(
	context.Context, ports.Scope,
) (ports.PrincipalSnapshot, error) {
	return ports.PrincipalSnapshot{PermissionRevision: 1}, nil
}

func (workAdoptionPrincipals) Exists(
	context.Context, ports.Scope, domain.Principal,
) (bool, int64, error) {
	return true, 1, nil
}

type workAdoptionTZDB struct{}

func (workAdoptionTZDB) Load(
	context.Context, string, string,
) (*time.Location, error) {
	return time.UTC, nil
}

type workAdoptionClock struct{}

func (workAdoptionClock) Now() time.Time {
	return time.Date(2026, 7, 19, 10, 0, 0, 0, time.UTC)
}

func workAdoptionService(transaction *sql.Tx,
	fixture legacyFixture,
) (*adoption.Service, workAdoptionScope, error) {
	database := workDatabaseAccess(transaction)
	tenantID := domain.ID(fixture.tenant)
	source, err := NewLegacySource(database, tenantID)
	if err != nil {
		return nil, workAdoptionScope{}, err
	}
	writer, err := NewCanonicalWriter(database, tenantID)
	if err != nil {
		return nil, workAdoptionScope{}, err
	}
	target, err := NewTargetStore(database, tenantID, writer)
	if err != nil {
		return nil, workAdoptionScope{}, err
	}
	service, err := adoption.NewService(adoption.Dependencies{Source: source, Target: target,
		Authorizer: workAdoptionAuthorizer{}, Forms: workAdoptionForms{},
		Workflow: workAdoptionWorkflow{}, Principals: workAdoptionPrincipals{},
		TZDB: workAdoptionTZDB{}, TZDBVersion: "2026a", Clock: workAdoptionClock{}})
	return service, workAdoptionScope{tenant: tenantID, actor: domain.ID(fixture.creator)}, err
}

func workDomainID(seed byte) domain.ID {
	return domain.ID(workUUID(90, seed))
}
