package aiinferencev1

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"

	canonicaljsonv1 "github.com/Onurlulardan/harpy/contracts/canonicaljson/v1"
)

type hashBinding struct {
	ContractVersion      string `json:"contract_version"`
	InvocationID         string `json:"invocation_id"`
	ReceiptID            string `json:"receipt_id"`
	TenantID             string `json:"tenant_id"`
	ConfigID             string `json:"config_id"`
	ConfigRevision       int64  `json:"config_revision"`
	CredentialRevision   int64  `json:"credential_revision"`
	AuthorizationVersion int64  `json:"authorization_version"`
	Operation            string `json:"operation"`
	Provider             string `json:"provider"`
	BaseURL              string `json:"base_url,omitempty"`
	Model                string `json:"model"`
}

type generationHashDocument struct {
	hashBinding
	Messages []Message          `json:"messages"`
	Tools    []ToolDefinition   `json:"tools,omitempty"`
	Settings GenerationSettings `json:"settings"`
}

type embeddingHashDocument struct {
	hashBinding
	Inputs     []string `json:"inputs"`
	Dimensions int      `json:"dimensions,omitempty"`
	TimeoutMS  int      `json:"timeout_ms"`
}

func CanonicalGenerationHash(request GenerationRequest) (string, error) {
	return digest(generationHashDocument{hashBinding: request.hashBinding(),
		Messages: request.Messages, Tools: request.Tools, Settings: request.Settings})
}

func CanonicalEmbeddingHash(request EmbeddingRequest) (string, error) {
	return digest(embeddingHashDocument{hashBinding: request.hashBinding(),
		Inputs: request.Inputs, Dimensions: request.Dimensions, TimeoutMS: request.TimeoutMS})
}

func CredentialEnvelopeDigest(document []byte) (string, error) {
	if len(document) == 0 {
		return "", nil
	}
	canonical, err := canonicaljsonv1.CanonicalizeObject(document, 16<<10)
	if err != nil {
		return "", ErrInvalidRequest
	}
	sum := sha256.Sum256(canonical)
	return hex.EncodeToString(sum[:]), nil
}

func HashMatches(actual, expected string) bool {
	return len(actual) == 64 && len(expected) == 64 &&
		subtle.ConstantTimeCompare([]byte(actual), []byte(expected)) == 1
}

func (binding Binding) hashBinding() hashBinding {
	return hashBinding{ContractVersion: binding.ContractVersion,
		InvocationID: binding.InvocationID, ReceiptID: binding.ReceiptID,
		TenantID: binding.TenantID, ConfigID: binding.ConfigID,
		ConfigRevision: binding.ConfigRevision, CredentialRevision: binding.CredentialRevision,
		AuthorizationVersion: binding.AuthorizationVersion, Operation: binding.Operation,
		Provider: binding.Provider, BaseURL: binding.BaseURL, Model: binding.Model}
}

func digest(value any) (string, error) {
	document, err := json.Marshal(value)
	if err != nil {
		return "", ErrInvalidRequest
	}
	canonical, err := canonicaljsonv1.CanonicalizeObject(document, MaximumRequestBodyBytes)
	clear(document)
	if err != nil {
		return "", ErrInvalidRequest
	}
	sum := sha256.Sum256(canonical)
	clear(canonical)
	return hex.EncodeToString(sum[:]), nil
}
