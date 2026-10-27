package v1adoption

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/work/domain"
	"github.com/Onurlulardan/harpy/core/internal/modules/work/ports"
)

type fixtureScope struct{ tenant, actor domain.ID }

func (scope fixtureScope) TenantID() domain.ID { return scope.tenant }
func (scope fixtureScope) ActorID() domain.ID  { return scope.actor }
func (fixtureScope) OperationID() string       { return "operation-1" }
func (fixtureScope) SnapshotQuery() bool       { return false }

type fixtureAuthorizer struct{}

func (fixtureAuthorizer) Authorize(context.Context, ports.Scope, string, domain.ID) error { return nil }

type fixtureSource struct{ records []SourceRecord }

func (source fixtureSource) LoadLegacy(context.Context, ports.Scope) ([]SourceRecord, error) {
	return append([]SourceRecord(nil), source.records...), nil
}

type fixtureTarget struct {
	plan     Plan
	evidence Evidence
}

func (target *fixtureTarget) PersistPlan(_ context.Context, _ ports.Scope, plan Plan) (Receipt, error) {
	target.plan = plan
	state := "RECONCILING"
	if len(plan.Issues) > 0 {
		state = "BLOCKED"
	}
	return Receipt{RunID: plan.RunID, State: state, SourceCount: plan.SourceCount,
		TargetCount: int64(len(plan.Records)), SHA256: targetDigest(plan.Records)}, nil
}
func (target *fixtureTarget) ComparePlan(_ context.Context, _ ports.Scope, plan Plan) (Evidence, error) {
	target.plan = plan
	return Evidence{RunID: plan.RunID, SourceCount: plan.SourceCount,
		TargetCount: int64(len(plan.Records)), SourceSHA256: plan.SourceSHA256,
		TargetSHA256: targetDigest(plan.Records), Accepted: len(plan.Issues) == 0,
		WriterFenced: true}, nil
}
func (target *fixtureTarget) LoadEvidence(context.Context, ports.Scope, domain.ID) (Evidence, error) {
	return target.evidence, nil
}
func (*fixtureTarget) RevertPlan(_ context.Context, _ ports.Scope, request ReverseRequest,
	_ Evidence,
) (Receipt, error) {
	return Receipt{RunID: request.RunID, State: "REVERSED"}, nil
}
func (*fixtureTarget) SealLegacy(_ context.Context, _ ports.Scope, request RetireRequest,
	_ Evidence,
) (Receipt, error) {
	return Receipt{RunID: request.RunID, State: "RETIRED"}, nil
}

type fixtureForms struct{}

func (fixtureForms) ResolveForm(context.Context, ports.Scope, string, *domain.ID,
	json.RawMessage,
) (FormReference, error) {
	return FormReference{ID: fixtureID(8), VersionID: fixtureID(9),
		PermissionRevision: 1}, nil
}

func (fixtureForms) RegisterRead(_ context.Context, _ ports.Scope,
	command ports.ResourceReferenceCommand,
) (ports.ResourceReferenceReceipt, error) {
	return ports.ResourceReferenceReceipt{ReferenceID: command.Link.ID,
		OperationID: command.OperationID, TargetID: command.Link.TargetID,
		TargetVersionID: command.Link.TargetVersionID, ReferenceEpoch: 1}, nil
}

type fixtureWorkflow struct{}

func (fixtureWorkflow) ResolveTask(context.Context, ports.Scope, domain.ID) (WorkflowProjection, error) {
	return WorkflowProjection{TaskID: fixtureID(10), VersionID: fixtureID(11),
		Revision: 1, PermissionRevision: 1, Summary: "approval",
		VisibleTo: []domain.Principal{{Kind: domain.PrincipalUser, ID: fixtureID(1)}}}, nil
}

type fixturePrincipals struct{}

func (fixturePrincipals) Resolve(context.Context, ports.Scope) (ports.PrincipalSnapshot, error) {
	return ports.PrincipalSnapshot{PermissionRevision: 1}, nil
}
func (fixturePrincipals) Exists(context.Context, ports.Scope, domain.Principal) (bool, int64, error) {
	return true, 1, nil
}

type fixtureTZDB struct{}

func (fixtureTZDB) Load(_ context.Context, name, version string) (*time.Location, error) {
	return time.LoadLocation(name)
}

type fixtureClock struct{}

func (fixtureClock) Now() time.Time { return time.Date(2026, 7, 19, 12, 0, 0, 0, time.UTC) }

func fixtureID(seed byte) domain.ID {
	var id domain.ID
	id[0], id[6], id[8], id[15] = 1, 0x70, 0x80, seed
	return id
}
func fixtureRequest() ApplyRequest {
	var digest [32]byte
	digest[0] = 1
	return ApplyRequest{
		RunID: fixtureID(1), Generation: 10, ManifestSHA256: digest}
}
func fixtureRecord(kind SourceKind, id domain.ID, value any) SourceRecord {
	document, _ := json.Marshal(value)
	return SourceRecord{Kind: kind, ID: id, Document: document, SHA256: sha256.Sum256(document)}
}

func fixtureService(records []SourceRecord, target *fixtureTarget) *Service {
	service, _ := NewService(Dependencies{Source: fixtureSource{records}, Target: target,
		Authorizer: fixtureAuthorizer{}, Forms: fixtureForms{}, Workflow: fixtureWorkflow{},
		Principals: fixturePrincipals{}, TZDB: fixtureTZDB{}, TZDBVersion: "2026a",
		Clock: fixtureClock{}})
	return service
}

func fixtureTenantScope() fixtureScope { return fixtureScope{fixtureID(2), fixtureID(3)} }
