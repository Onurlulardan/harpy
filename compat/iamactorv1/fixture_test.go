package iamactorv1

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"errors"
	"time"

	actorassertionv1 "github.com/Onurlulardan/harpy/contracts/actorassertion/v1"
)

type consumerFixture struct {
	now         time.Time
	privateKey  ed25519.PrivateKey
	publicKey   ed25519.PublicKey
	wire        string
	expectation actorassertionv1.Expectation
}

func newConsumerFixture(t testingT) consumerFixture {
	t.Helper()
	now := time.Unix(2_000_000_000, 0).UTC()
	privateKey := ed25519.NewKeyFromSeed(bytes.Repeat([]byte{41}, ed25519.SeedSize))
	publicKey := append(ed25519.PublicKey(nil), privateKey[ed25519.SeedSize:]...)
	target, err := actorassertionv1.TargetBindingSHA256(
		"POST", "/api/v1/forms/records", "record-17",
	)
	if err != nil {
		t.Fatal(err)
	}
	decision := actorassertionv1.DecisionEvidence{
		TenantID:     fixtureUUID(0x10, 0x01),
		SubjectID:    fixtureUUID(0x20, 0x02),
		PermissionID: fixtureUUID(0x30, 0x03),
		TargetDomain: actorassertionv1.DomainDocuments,
		ResourceKey:  "documents.items", ActionKey: "create",
		TargetBindingSHA256: target, IAMGroupGeneration: 11,
		IAMGeneration: 12, TargetGeneration: 13, TokenVersion: 7,
		AuthorizationVersion: 19, AuthorizationClock: 23,
	}
	signer, err := actorassertionv1.NewSigner(
		"edge-2026-07", privateKey, bytes.NewReader(bytes.Repeat([]byte{5}, 16)),
	)
	if err != nil {
		t.Fatal(err)
	}
	credential, err := signer.Issue(decision, now)
	signer.Close()
	if err != nil {
		t.Fatal(err)
	}
	wire, err := credential.Wire()
	if err != nil {
		t.Fatal(err)
	}
	expectation := actorassertionv1.Expectation{TenantID: decision.TenantID,
		TargetDomain: decision.TargetDomain, ResourceKey: decision.ResourceKey,
		ActionKey: decision.ActionKey, TargetBindingSHA256: decision.TargetBindingSHA256,
		IAMGroupGeneration: decision.IAMGroupGeneration,
		IAMGeneration:      decision.IAMGeneration, TargetGeneration: decision.TargetGeneration}
	return consumerFixture{now: now, privateKey: privateKey, publicKey: publicKey,
		wire: wire, expectation: expectation}
}

func fixtureUUID(first, last byte) [16]byte {
	value := [16]byte{first}
	value[6] = 0x40
	value[8] = 0x80
	value[15] = last
	return value
}

type testingT interface {
	Helper()
	Fatal(...any)
}

type transactionStub struct {
	row       Row
	statement string
	arguments []any
	order     *[]string
}

func (transaction *transactionStub) QueryRow(
	_ context.Context,
	statement string,
	arguments ...any,
) Row {
	transaction.statement = statement
	transaction.arguments = append([]any(nil), arguments...)
	if transaction.order != nil {
		*transaction.order = append(*transaction.order, "consume")
	}
	return transaction.row
}

type rowStub struct {
	consumedAt time.Time
	err        error
}

func (row rowStub) Scan(destinations ...any) error {
	if row.err != nil {
		return row.err
	}
	if len(destinations) != 1 {
		return errors.New("unexpected scan")
	}
	*destinations[0].(*time.Time) = row.consumedAt
	return nil
}

type stateError string

func (failure stateError) Error() string    { return string(failure) }
func (failure stateError) SQLState() string { return string(failure) }
