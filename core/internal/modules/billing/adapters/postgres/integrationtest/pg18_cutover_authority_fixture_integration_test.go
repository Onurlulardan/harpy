//go:build integration

package integrationtest

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

var (
	billingCutoverFixtureRequester = uuid.MustParse(
		"01900000-0000-7000-8000-000000000091",
	)
	billingCutoverFixtureApprover = uuid.MustParse(
		"01900000-0000-7000-8000-000000000092",
	)
	billingCutoverFixtureChangedAt = time.Date(
		2026, time.July, 29, 0, 0, 0, 0, time.UTC,
	)
)

type billingCutoverFixtureExecutor interface {
	ExecContext(context.Context, string, ...any) (sql.Result, error)
}

type billingCutoverAuthorityFixture struct {
	scopeKind, groupID, state, writer string
	scopeID                           uuid.UUID
	generation                        int64
	domains                           []string
	requestedBy, approvedBy           uuid.UUID
	reason, changeReference           string
	acceptanceDigest                  any
	changedAt                         time.Time
}

func newBillingCutoverAuthorityFixture(
	scopeKind string,
	scopeID uuid.UUID,
	groupID string,
	generation int64,
	state string,
	writer string,
	acceptanceDigest any,
	domains ...string,
) billingCutoverAuthorityFixture {
	return billingCutoverAuthorityFixture{
		scopeKind: scopeKind, scopeID: scopeID, groupID: groupID,
		generation: generation, state: state, writer: writer,
		domains:          append([]string(nil), domains...),
		requestedBy:      billingCutoverFixtureRequester,
		approvedBy:       billingCutoverFixtureApprover,
		reason:           "billing authority cutover",
		changeReference:  "billing-cutover:test",
		acceptanceDigest: acceptanceDigest,
		changedAt:        billingCutoverFixtureChangedAt,
	}
}

func seedBillingCutoverAuthority(
	ctx context.Context,
	executor billingCutoverFixtureExecutor,
	fixture billingCutoverAuthorityFixture,
) error {
	_, err := executor.ExecContext(ctx, `
INSERT INTO platform.cutover_groups(
 scope_kind,scope_id,group_id,state,writer,generation,requested_by,approved_by,
 reason,change_reference,acceptance_digest,changed_at
) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)`,
		fixture.scopeKind, fixture.scopeID, fixture.groupID,
		fixture.state, fixture.writer, fixture.generation,
		fixture.requestedBy, fixture.approvedBy, fixture.reason,
		fixture.changeReference, fixture.acceptanceDigest, fixture.changedAt)
	if err != nil {
		return fmt.Errorf("insert cutover group %s: %w", fixture.groupID, err)
	}
	for _, domain := range fixture.domains {
		if err = seedBillingDomainCutover(
			ctx, executor, fixture, domain,
		); err != nil {
			return err
		}
	}
	for index, domain := range fixture.domains {
		_, err = executor.ExecContext(ctx, `
INSERT INTO platform.cutover_group_members(
 scope_kind,scope_id,group_id,ordinal,domain
) VALUES($1,$2,$3,$4,$5)`,
			fixture.scopeKind, fixture.scopeID, fixture.groupID, index+1, domain)
		if err != nil {
			return fmt.Errorf("insert cutover member %s: %w", domain, err)
		}
	}
	return nil
}

func seedBillingDomainCutover(
	ctx context.Context,
	executor billingCutoverFixtureExecutor,
	fixture billingCutoverAuthorityFixture,
	domain string,
) error {
	_, err := executor.ExecContext(ctx, `
INSERT INTO platform.domain_cutovers(
 scope_kind,scope_id,domain,state,writer,generation,requested_by,approved_by,
 reason,change_reference,acceptance_digest,changed_at
) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)`,
		fixture.scopeKind, fixture.scopeID, domain,
		fixture.state, fixture.writer, fixture.generation,
		fixture.requestedBy, fixture.approvedBy, fixture.reason,
		fixture.changeReference, fixture.acceptanceDigest, fixture.changedAt)
	if err != nil {
		return fmt.Errorf("insert domain cutover %s: %w", domain, err)
	}
	return nil
}

func insertBillingTenantTopology(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
	tenantID uuid.UUID,
	generation int64,
	digest string,
) {
	t.Helper()
	schema := "t_" + strings.ReplaceAll(tenantID.String()[:12], "-", "")
	transaction, err := database.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer transaction.Rollback()
	if _, err = transaction.ExecContext(ctx, `
INSERT INTO platform.tenants(id,schema_name,status,deleted_at)
VALUES($1,$2,'ACTIVE',NULL)`, tenantID, schema); err != nil {
		t.Fatal(err)
	}
	fixture := newBillingCutoverAuthorityFixture(
		"tenant", tenantID, "billing_forms", generation,
		"v2", "v2", digest, "billing",
	)
	if err = seedBillingCutoverAuthority(ctx, transaction, fixture); err != nil {
		t.Fatal(err)
	}
	if err = transaction.Commit(); err != nil {
		t.Fatal(err)
	}
}
