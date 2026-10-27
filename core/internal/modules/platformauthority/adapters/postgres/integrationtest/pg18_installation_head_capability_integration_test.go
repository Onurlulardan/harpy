//go:build integration

package postgres_test

import (
	"context"
	"database/sql"
	"strings"
	"testing"
	"time"
)

type authorityHeadFixture struct {
	activationOperationID, authorityOperationID    string
	bindingID, environmentID, tenantID, schemaName string
	stateSHA256, proposalSHA256, ackSHA256         string
	manifestSHA256                                 string
}

type authorityHeadState struct {
	status                string
	pins, heads, receipts int
}

func TestInstallationHeadCapabilityAgainstPostgreSQL184(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	database := startPlatformAuthorityDatabase(t, ctx)
	resetPlatformSchema(t, ctx, database)
	applyAuthorityFreshDocuments(t, ctx, database)
	prepareAuthorityActivationStepFixture(t, ctx, database)
	exerciseInstallationHeadCapability(t, ctx, database)
}

func exerciseInstallationHeadCapability(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
) {
	t.Helper()
	fixture := createAuthorityHeadFixture(t, ctx, database)
	initial := readAuthorityHeadState(t, ctx, database, fixture)
	correct := authorityActivationCapability{
		operationID:   fixture.activationOperationID,
		commandKey:    "platformauthority.pin_protected_authority_head",
		environmentID: fixture.environmentID, tenantID: fixture.tenantID,
		schemaName:    fixture.schemaName,
		operationKind: "saas_customer_tenant_provision",
	}
	for _, attempt := range []struct {
		name       string
		capability *authorityActivationCapability
	}{
		{name: "no_capability"},
		{name: "wrong_environment", capability: changedCapability(
			correct, "environment", "ffffffff-ffff-4fff-8fff-ffffffffffff",
		)},
		{name: "wrong_kind", capability: changedCapability(
			correct, "kind", "environment_initialize",
		)},
	} {
		t.Run(attempt.name, func(t *testing.T) {
			_, err := callInstallationHeadPin(
				t, ctx, database, "platformauthority_installation_head_pin_v3",
				"harpy-v2:core-worker", fixture, attempt.capability,
			)
			if authorityPGCode(err) != "42501" {
				t.Fatalf("%s code=%q err=%v",
					attempt.name, authorityPGCode(err), err)
			}
			requireAuthorityHeadState(t, ctx, database, fixture, initial)
		})
	}
	replayed, err := callInstallationHeadPin(
		t, ctx, database, "platformauthority_installation_head_pin_v3",
		"harpy-v2:core-worker", fixture, &correct,
	)
	if err != nil || replayed {
		t.Fatalf("customer worker pin replay=%t err=%v", replayed, err)
	}
	committed := authorityHeadState{
		status: "completed", pins: 1, heads: 1, receipts: 1,
	}
	requireAuthorityHeadState(t, ctx, database, fixture, committed)
	replayed, err = callInstallationHeadPin(
		t, ctx, database, "platformauthority_installation_head_pin_v3",
		"harpy-v2:core-worker", fixture, &correct,
	)
	if err != nil || !replayed {
		t.Fatalf("customer worker replay=%t err=%v", replayed, err)
	}
	replayed, err = callInstallationHeadPin(
		t, ctx, database, "platformauthority_installation_head_pin_v2",
		"harpy-v2:core-operator-initialize", fixture, nil,
	)
	if replayed || authorityPGCode(err) != "42501" {
		t.Fatalf(
			"retired generation-23 v2 replay=%t code=%q err=%v",
			replayed, authorityPGCode(err), err,
		)
	}
	requireAuthorityHeadState(t, ctx, database, fixture, committed)
}

func changedCapability(
	value authorityActivationCapability,
	field, replacement string,
) *authorityActivationCapability {
	switch field {
	case "environment":
		value.environmentID = replacement
	case "kind":
		value.operationKind = replacement
	}
	return &value
}

func createAuthorityHeadFixture(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
) authorityHeadFixture {
	t.Helper()
	now := time.Date(2026, 7, 29, 10, 32, 0, 0, time.UTC)
	value := authorityHeadFixture{
		activationOperationID: "11111111-1111-4111-8111-111111111111",
		authorityOperationID:  "22222222-2222-4222-8222-222222222222",
		bindingID:             "33333333-3333-4333-8333-333333333333",
		environmentID:         "44444444-4444-4444-8444-444444444444",
		tenantID:              "55555555-5555-4555-8555-555555555555",
		schemaName:            "t_aaaaaaaaaaaaaaaaaaaaaaaaaa",
		stateSHA256:           strings.Repeat("a", 64),
		proposalSHA256:        strings.Repeat("b", 64),
		ackSHA256:             strings.Repeat("c", 64),
		manifestSHA256:        strings.Repeat("d", 64),
	}
	if _, err := database.ExecContext(ctx, bindSQL,
		value.bindingID, value.environmentID, "saas", "saas_provider",
		nil, nil, nil, int64(1), "proposal.v1", strings.Repeat("e", 64),
		nil, nil, now); err != nil {
		t.Fatal(err)
	}
	if _, err := database.ExecContext(ctx, prepareSQL,
		value.authorityOperationID, value.bindingID, "initial_tenant_authority",
		"tenant", value.tenantID, int64(1), nil, value.stateSHA256,
		[]byte("state"), value.proposalSHA256,
		[]byte(`{"signer_kid":"proposal.v1"}`), nil,
		now, now.Add(10*time.Minute), "proposal.v1", now); err != nil {
		t.Fatal(err)
	}
	if _, err := database.ExecContext(ctx, recordAckSQL,
		value.authorityOperationID, value.bindingID, "tenant", value.tenantID,
		int64(1), nil, value.stateSHA256, value.proposalSHA256,
		value.ackSHA256, []byte("ack"), int64(1), now); err != nil {
		t.Fatal(err)
	}
	return value
}
