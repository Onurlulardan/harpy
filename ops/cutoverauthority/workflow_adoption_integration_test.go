//go:build integration

package main

import (
	"bytes"
	"context"
	"os"
	"testing"
	"time"

	cutoverv1 "github.com/Onurlulardan/harpy/contracts/cutover/v1"
)

func TestAuthorityAdoptionSurvivesQuiescenceAndPostCommitPublicationCrashesPG184(
	t *testing.T,
) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	database := startWorkflowDatabase(t, ctx)
	directory := t.TempDir()
	now := time.Now().UTC().Truncate(time.Second)
	fixture := prepareWorkflowAuthorityArtifacts(t, ctx, database, directory, now)
	config := fixture.adoptionInvocation()
	bundle, err := loadAuthorityBundle(fixture.paths, time.Now().UTC())
	if err != nil {
		t.Fatalf("load signed adoption bundle: %v", err)
	}
	signer, err := loadAuthoritySigner(fixture.signer.signerPath, bundle.root)
	if err != nil {
		t.Fatal(err)
	}
	first, err := openAuthorityDatabase(ctx, fixture.rootDSNPath, "test")
	if err != nil {
		t.Fatal(err)
	}
	policy := bundle.operation.Value().QuiescencePolicy.Roles[0]
	if _, err := quiesceMigrator(ctx, first, policy); err != nil {
		t.Fatalf("prepare simulated post-quiescence crash: %v", err)
	}
	_ = first.Close(ctx)
	second, err := openAuthorityDatabase(ctx, fixture.rootDSNPath, "test")
	if err != nil {
		t.Fatal(err)
	}
	quiescence, err := quiesceMigrator(ctx, second, policy)
	if err != nil {
		t.Fatalf("resume exact quiescence state: %v", err)
	}
	committed, err := adoptInTransaction(ctx, second, bundle, signer, quiescence)
	if err != nil {
		t.Fatalf("commit adoption before simulated publication crash: %v", err)
	}
	_ = second.Close(ctx)
	if _, err := os.Stat(fixture.receiptPath); !os.IsNotExist(err) {
		t.Fatal("receipt was published before explicit replay")
	}
	result, err := executeAuthorityAdoption(ctx, config)
	if err != nil {
		t.Fatalf("replay committed adoption: %v", err)
	}
	if !result.Replayed || result.EvidenceSHA256 != committed.digest {
		t.Fatalf("unexpected committed replay result: %+v", result)
	}
	assertWorkflowPublishedAdoption(t, fixture, committed)
	result, err = executeAuthorityAdoption(ctx, config)
	if err != nil || !result.Replayed {
		t.Fatalf("repeat identical adoption replay: result=%+v err=%v", result, err)
	}
	assertWorkflowPublishedAdoption(t, fixture, committed)
	assertWorkflowAdoptionState(t, ctx, database, policy)
}

func assertWorkflowPublishedAdoption(
	t *testing.T, fixture workflowAuthorityFixture, expected signedAdoptionReceipt,
) {
	t.Helper()
	receipt, err := os.ReadFile(fixture.receiptPath)
	if err != nil {
		t.Fatal(err)
	}
	signature, err := os.ReadFile(fixture.signaturePath)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(receipt, expected.document) ||
		!bytes.Equal(signature, expected.signatureDocument) {
		t.Fatal("published replay bytes differ from committed inventory")
	}
}

func assertWorkflowAdoptionState(
	t *testing.T, ctx context.Context, database workflowDatabase,
	policy cutoverv1.AuthorityQuiescedRolePolicy,
) {
	t.Helper()
	var inventory int
	if err := database.admin.QueryRow(ctx,
		"SELECT count(*)::integer FROM cutover_authority.authority_adoption_inventory",
	).Scan(&inventory); err != nil || inventory != 1 {
		t.Fatalf("unexpected adoption inventory: count=%d err=%v", inventory, err)
	}
	observed, err := readQuiescedRole(ctx, database.admin, policy.Principal)
	if err != nil || observed.state != policy.PostState {
		t.Fatalf("migrator was not restored: state=%+v err=%v", observed.state, err)
	}
	for _, identity := range workflowAuthorityObjects {
		state, err := readCatalogState(ctx, database.admin, identity.kind,
			identity.schema, identity.name, identity.arguments)
		if err != nil || state.owner != "harpy_cutover_authority_owner" {
			t.Fatalf("authority owner differs for %s: owner=%s err=%v",
				identity.name, state.owner, err)
		}
	}
}
