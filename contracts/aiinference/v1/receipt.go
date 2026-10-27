package aiinferencev1

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"errors"
	"time"

	canonicaljsonv1 "github.com/Onurlulardan/harpy/contracts/canonicaljson/v1"
	"github.com/google/uuid"
)

const receiptMACDomain = "harpy:ai-inference:usage-receipt:v1\x00"

var ErrInvalidReceipt = errors.New("AI inference usage receipt is invalid")

type ReceiptExpected struct {
	ReceiptID, InvocationID, RequestHash string
	Provider, ResolvedModel              string
}

type ReceiptSigner struct {
	key      []byte
	keyID    string
	revision int64
}

func NewReceiptSigner(key []byte, keyID string, revision int64) (*ReceiptSigner, error) {
	if len(key) < 32 || !validName(keyID) || revision < 1 {
		return nil, ErrInvalidReceipt
	}
	return &ReceiptSigner{key: append([]byte(nil), key...), keyID: keyID,
		revision: revision}, nil
}

func (signer *ReceiptSigner) Sign(receipt UsageReceipt) (SignedReceipt, error) {
	receipt.ContractVersion = ReceiptContractVersion
	if signer == nil || len(signer.key) < 32 || !validReceipt(receipt) {
		return SignedReceipt{}, ErrInvalidReceipt
	}
	payload, err := receiptBytes(receipt)
	if err != nil {
		return SignedReceipt{}, err
	}
	mac := hmac.New(sha256.New, signer.key)
	_, _ = mac.Write([]byte(receiptMACDomain))
	_, _ = mac.Write(payload)
	signature := mac.Sum(nil)
	clear(payload)
	encoded := base64.RawURLEncoding.EncodeToString(signature)
	clear(signature)
	return SignedReceipt{ContractVersion: SignedReceiptContractVersion,
		Algorithm: "hmac-sha256", SigningKeyID: signer.keyID,
		SigningKeyRevision: signer.revision, Receipt: receipt, Signature: encoded}, nil
}

func (signer *ReceiptSigner) Verify(signed SignedReceipt, expected ReceiptExpected) error {
	if signer == nil || len(signer.key) < 32 || signed.ContractVersion != SignedReceiptContractVersion ||
		signed.Algorithm != "hmac-sha256" || signed.SigningKeyID != signer.keyID ||
		signed.SigningKeyRevision != signer.revision || !validReceipt(signed.Receipt) ||
		!expected.matches(signed.Receipt) {
		return ErrInvalidReceipt
	}
	actual, err := base64.RawURLEncoding.DecodeString(signed.Signature)
	if err != nil || len(actual) != sha256.Size ||
		base64.RawURLEncoding.EncodeToString(actual) != signed.Signature {
		clear(actual)
		return ErrInvalidReceipt
	}
	defer clear(actual)
	payload, err := receiptBytes(signed.Receipt)
	if err != nil {
		return err
	}
	defer clear(payload)
	mac := hmac.New(sha256.New, signer.key)
	_, _ = mac.Write([]byte(receiptMACDomain))
	_, _ = mac.Write(payload)
	want := mac.Sum(nil)
	defer clear(want)
	if subtle.ConstantTimeCompare(actual, want) != 1 {
		return ErrInvalidReceipt
	}
	return nil
}

func (signer *ReceiptSigner) Close() {
	if signer != nil {
		clear(signer.key)
		signer.key = nil
	}
}

func (expected ReceiptExpected) matches(receipt UsageReceipt) bool {
	return expected.ReceiptID == receipt.ReceiptID &&
		expected.InvocationID == receipt.InvocationID &&
		HashMatches(expected.RequestHash, receipt.RequestHash) &&
		expected.Provider == receipt.Provider && expected.ResolvedModel == receipt.ResolvedModel
}

func validReceipt(receipt UsageReceipt) bool {
	receiptID, receiptErr := uuid.Parse(receipt.ReceiptID)
	invocation, invocationErr := uuid.Parse(receipt.InvocationID)
	return receipt.ContractVersion == ReceiptContractVersion && receiptErr == nil &&
		invocationErr == nil && receiptID.Version() == 7 && invocation.Version() == 7 &&
		hexPattern.MatchString(receipt.RequestHash) && validProvider(receipt.Provider) &&
		modelPattern.MatchString(receipt.ResolvedModel) && ValidFinishClass(receipt.FinishClass) &&
		receipt.Usage.InputTokens >= 0 && receipt.Usage.OutputTokens >= 0 &&
		receipt.Usage.CachedTokens >= 0 && receipt.Usage.CachedTokens <= receipt.Usage.InputTokens &&
		receipt.Usage.InputTokens <= MaximumReportedTokens &&
		receipt.Usage.OutputTokens <= MaximumReportedTokens &&
		receipt.Usage.CachedTokens <= MaximumReportedTokens &&
		receipt.DurationMS >= 0 && receipt.DurationMS <= MaximumTimeoutMS+1000 &&
		(receipt.ProviderReceiptHash == "" || hexPattern.MatchString(receipt.ProviderReceiptHash)) &&
		!receipt.IssuedAt.IsZero() && receipt.IssuedAt.Location() == time.UTC
}

func ValidFinishClass(value string) bool {
	return value == FinishStop || value == FinishLength ||
		value == FinishToolProposals || value == FinishContentFilter || value == FinishEmbedding
}

func receiptBytes(receipt UsageReceipt) ([]byte, error) {
	document, err := json.Marshal(receipt)
	if err != nil {
		return nil, ErrInvalidReceipt
	}
	canonical, err := canonicaljsonv1.CanonicalizeObject(document, 16<<10)
	clear(document)
	if err != nil {
		return nil, ErrInvalidReceipt
	}
	return canonical, nil
}
