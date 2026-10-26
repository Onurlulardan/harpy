package iamactorv1

import (
	"context"
	"crypto/ed25519"
	"errors"
	"net/http"
	"strings"
	"testing"
	"time"

	actorassertionv1 "github.com/Onurlulardan/harpy/contracts/actorassertion/v1"
)

func TestConsumeVerifiesThenUsesOneCallerOwnedTransaction(t *testing.T) {
	fixture := newConsumerFixture(t)
	consumer, err := NewConsumer(map[string]ed25519.PublicKey{
		"edge-2026-07": fixture.publicKey,
	}, func() time.Time { return fixture.now })
	if err != nil {
		t.Fatal(err)
	}
	order := []string{}
	transaction := &transactionStub{row: rowStub{consumedAt: fixture.now}, order: &order}
	receipt, err := consumer.Consume(
		context.Background(), transaction, fixture.wire, fixture.expectation,
		func(_ context.Context, used Transaction, claims actorassertionv1.Claims) error {
			order = append(order, "effect")
			if used != transaction || claims.Decision.TenantID != fixture.expectation.TenantID {
				t.Fatal("effect did not receive the verified transaction and claims")
			}
			return nil
		},
	)
	if err != nil || !receipt.Valid() || strings.Join(order, ",") != "consume,effect" {
		t.Fatalf("receipt=%v order=%v err=%v", receipt.Valid(), order, err)
	}
	if transaction.statement != consumeStatement || len(transaction.arguments) != 17 {
		t.Fatalf("consume invocation differs: args=%d", len(transaction.arguments))
	}
}

func TestConsumeFailsBeforeEffectAndClassifiesReplay(t *testing.T) {
	fixture := newConsumerFixture(t)
	consumer, _ := NewConsumer(map[string]ed25519.PublicKey{
		"edge-2026-07": fixture.publicKey,
	}, func() time.Time { return fixture.now })
	called := false
	transaction := &transactionStub{row: rowStub{err: stateError("P5604")}}
	_, err := consumer.Consume(
		context.Background(), transaction, fixture.wire, fixture.expectation,
		func(context.Context, Transaction, actorassertionv1.Claims) error {
			called = true
			return nil
		},
	)
	if !Is(err, CodeReplay) || called {
		t.Fatalf("replay err=%v effect_called=%t", err, called)
	}
	wrong := fixture.expectation
	wrong.TargetGeneration++
	transaction.statement = ""
	if _, err := consumer.Consume(
		context.Background(), transaction, fixture.wire, wrong,
		func(context.Context, Transaction, actorassertionv1.Claims) error { return nil },
	); !Is(err, CodeUnauthorized) || transaction.statement != "" {
		t.Fatalf("wrong binding reached database: err=%v", err)
	}
}

func TestConsumeHeaderRejectsDuplicatesAndRemovesCredential(t *testing.T) {
	fixture := newConsumerFixture(t)
	consumer, _ := NewConsumer(map[string]ed25519.PublicKey{
		"edge-2026-07": fixture.publicKey,
	}, func() time.Time { return fixture.now })
	request, _ := http.NewRequest(http.MethodPost, "https://v1.invalid/records", nil)
	request.Header[actorassertionv1.HeaderName] = []string{fixture.wire}
	transaction := &transactionStub{row: rowStub{consumedAt: fixture.now}}
	_, err := consumer.ConsumeHeader(
		request, transaction, fixture.expectation,
		func(context.Context, Transaction, actorassertionv1.Claims) error {
			if len(request.Header) != 0 {
				t.Fatal("verified credential remained visible to the effect")
			}
			return nil
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	request.Header[actorassertionv1.HeaderName] = []string{fixture.wire, fixture.wire}
	if _, err := TakeAssertionHeader(request.Header); !Is(err, CodeInvalid) {
		t.Fatalf("duplicate assertion error = %v", err)
	}
}

func TestConsumeReturnsEffectFailureForCallerRollback(t *testing.T) {
	fixture := newConsumerFixture(t)
	consumer, _ := NewConsumer(map[string]ed25519.PublicKey{
		"edge-2026-07": fixture.publicKey,
	}, func() time.Time { return fixture.now })
	want := errors.New("business effect failed")
	transaction := &transactionStub{row: rowStub{consumedAt: fixture.now}}
	if _, err := consumer.Consume(
		context.Background(), transaction, fixture.wire, fixture.expectation,
		func(context.Context, Transaction, actorassertionv1.Claims) error { return want },
	); !errors.Is(err, want) {
		t.Fatalf("effect failure = %v", err)
	}
	consumer.Close()
	if _, err := consumer.Consume(
		context.Background(), transaction, fixture.wire, fixture.expectation,
		func(context.Context, Transaction, actorassertionv1.Claims) error { return nil },
	); !Is(err, CodeInvalid) {
		t.Fatalf("closed consumer error = %v", err)
	}
}
