package billingv1adoptionapp

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type fakeScopeStore struct {
	state     string
	applied   int64
	verified  Reconciliation
	rehearsal RollbackRehearsal
	document  []byte
	evidence  CutoverEvidence
}

func (store *fakeScopeStore) Register(
	_ context.Context,
	_ Operation,
	_ ScopePlan,
	document []byte,
	evidence CutoverEvidence,
) (string, error) {
	store.document = append([]byte(nil), document...)
	store.evidence = evidence
	return store.state, nil
}

func (store *fakeScopeStore) Apply(context.Context, Operation) (int64, error) {
	return store.applied, nil
}

func (store *fakeScopeStore) Verify(
	context.Context,
	Operation,
) (Reconciliation, error) {
	return store.verified, nil
}

func (store *fakeScopeStore) RehearseRollback(
	context.Context,
	Operation,
	uuid.UUID,
	string,
	time.Time,
) (RollbackRehearsal, error) {
	return store.rehearsal, nil
}

func scopeOperation(plan ScopePlan) Operation {
	_, digest, err := EncodeScopePlan(plan)
	if err != nil {
		panic(err)
	}
	return Operation{OperationID: uuid.New(), RunID: uuid.New(),
		AuthorityID: uuid.New(), PlanSHA256: digest}
}

func emptyEncodedPlan() ScopePlan {
	plan := emptyScopePlan("global", "70000000-0000-4000-8000-000000000007",
		"UTC", time.Date(2026, 7, 17, 10, 0, 0, 0, time.UTC))
	evidence, _ := categoryEvidence("plans", []string{}, 0)
	plan.Categories = append(plan.Categories, evidence)
	plan.SourceSHA256, _ = sourceRoot(plan.Categories)
	finalizePlan(&plan)
	return plan
}
