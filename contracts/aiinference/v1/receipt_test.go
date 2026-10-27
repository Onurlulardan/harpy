package aiinferencev1

import (
	"errors"
	"testing"
	"time"
)

func TestSignedReceiptUsesConstantTimeMACAndExactReplayBinding(t *testing.T) {
	signer, err := NewReceiptSigner([]byte("01234567890123456789012345678901"),
		ReceiptSigningKeyID, ReceiptSigningKeyRevision)
	if err != nil {
		t.Fatal(err)
	}
	defer signer.Close()
	receipt := UsageReceipt{ContractVersion: ReceiptContractVersion,
		ReceiptID:    "0198eac5-5ee9-7e0f-8c4a-8493a072f802",
		InvocationID: "0198eac5-5ee9-7e0f-8c4a-8493a072f801",
		RequestHash:  "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		Provider:     "openai", ResolvedModel: "gpt-5", Usage: TokenUsage{InputTokens: 3,
			OutputTokens: 2, CachedTokens: 1}, FinishClass: FinishStop, DurationMS: 21,
		IssuedAt: time.Date(2026, 8, 1, 12, 0, 0, 0, time.UTC)}
	signed, err := signer.Sign(receipt)
	if err != nil {
		t.Fatal(err)
	}
	expected := ReceiptExpected{ReceiptID: receipt.ReceiptID, InvocationID: receipt.InvocationID,
		RequestHash: receipt.RequestHash, Provider: receipt.Provider, ResolvedModel: receipt.ResolvedModel}
	if err := signer.Verify(signed, expected); err != nil {
		t.Fatal(err)
	}
	expected.InvocationID = "0198eac5-5ee9-7e0f-8c4a-8493a072f899"
	if err := signer.Verify(signed, expected); !errors.Is(err, ErrInvalidReceipt) {
		t.Fatalf("cross-invocation receipt error=%v", err)
	}
	replacement := byte('A')
	if signed.Signature[len(signed.Signature)-1] == replacement {
		replacement = 'B'
	}
	signed.Signature = signed.Signature[:len(signed.Signature)-1] + string(replacement)
	if err := signer.Verify(signed, ReceiptExpected{ReceiptID: receipt.ReceiptID,
		InvocationID: receipt.InvocationID, RequestHash: receipt.RequestHash,
		Provider: receipt.Provider, ResolvedModel: receipt.ResolvedModel}); !errors.Is(err, ErrInvalidReceipt) {
		t.Fatalf("tampered MAC error=%v", err)
	}
}

func TestReceiptRejectsUnboundedProviderUsage(t *testing.T) {
	signer, _ := NewReceiptSigner([]byte("01234567890123456789012345678901"),
		ReceiptSigningKeyID, ReceiptSigningKeyRevision)
	defer signer.Close()
	receipt := UsageReceipt{ReceiptID: "0198eac5-5ee9-7e0f-8c4a-8493a072f802",
		InvocationID: "0198eac5-5ee9-7e0f-8c4a-8493a072f801",
		RequestHash:  "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		Provider:     "openai", ResolvedModel: "gpt-5", FinishClass: FinishStop,
		Usage:    TokenUsage{InputTokens: MaximumReportedTokens + 1},
		IssuedAt: time.Date(2026, 8, 1, 12, 0, 0, 0, time.UTC)}
	if _, err := signer.Sign(receipt); !errors.Is(err, ErrInvalidReceipt) {
		t.Fatalf("unbounded usage error=%v", err)
	}
}
