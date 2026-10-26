package aiinferencev1

import (
	"encoding/json"
	"time"
)

type Binding struct {
	ContractVersion      string          `json:"contract_version"`
	InvocationID         string          `json:"invocation_id"`
	ReceiptID            string          `json:"receipt_id"`
	RequestHash          string          `json:"request_hash"`
	TenantID             string          `json:"tenant_id"`
	ConfigID             string          `json:"config_id"`
	ConfigRevision       int64           `json:"config_revision"`
	CredentialRevision   int64           `json:"credential_revision"`
	AuthorizationVersion int64           `json:"authorization_version"`
	Operation            string          `json:"operation"`
	Provider             string          `json:"provider"`
	BaseURL              string          `json:"base_url,omitempty"`
	Model                string          `json:"model"`
	CredentialEnvelope   json.RawMessage `json:"credential_envelope,omitempty"`
	SignedContext        json.RawMessage `json:"signed_context"`
}

type Message struct {
	Role          string         `json:"role"`
	Content       string         `json:"content,omitempty"`
	ToolProposals []ToolProposal `json:"tool_proposals,omitempty"`
	ToolResult    *ToolResult    `json:"tool_result,omitempty"`
}

type ToolDefinition struct {
	Name        string          `json:"name"`
	Description string          `json:"description,omitempty"`
	InputSchema json.RawMessage `json:"input_schema"`
}

type ToolProposal struct {
	ProposalID string          `json:"proposal_id"`
	Name       string          `json:"name"`
	Arguments  json.RawMessage `json:"arguments"`
}

type ToolResult struct {
	ProposalID string `json:"proposal_id"`
	Name       string `json:"name"`
	Content    string `json:"content"`
	IsError    bool   `json:"is_error"`
}

type GenerationSettings struct {
	MaximumOutputTokens int      `json:"maximum_output_tokens"`
	TimeoutMS           int      `json:"timeout_ms"`
	Temperature         *float64 `json:"temperature,omitempty"`
}

type GenerationRequest struct {
	Binding
	Messages []Message          `json:"messages"`
	Tools    []ToolDefinition   `json:"tools,omitempty"`
	Settings GenerationSettings `json:"settings"`
}

type EmbeddingRequest struct {
	Binding
	Inputs     []string `json:"inputs"`
	Dimensions int      `json:"dimensions,omitempty"`
	TimeoutMS  int      `json:"timeout_ms"`
}

type TokenUsage struct {
	InputTokens  int64 `json:"input_tokens"`
	OutputTokens int64 `json:"output_tokens"`
	CachedTokens int64 `json:"cached_tokens"`
}

type GenerationResult struct {
	ContractVersion string         `json:"contract_version"`
	InvocationID    string         `json:"invocation_id"`
	RequestHash     string         `json:"request_hash"`
	Provider        string         `json:"provider"`
	ResolvedModel   string         `json:"resolved_model"`
	Content         string         `json:"content"`
	ToolProposals   []ToolProposal `json:"tool_proposals"`
	FinishClass     string         `json:"finish_class"`
	UsageReceipt    SignedReceipt  `json:"usage_receipt"`
}

type EmbeddingResult struct {
	ContractVersion string        `json:"contract_version"`
	InvocationID    string        `json:"invocation_id"`
	RequestHash     string        `json:"request_hash"`
	Provider        string        `json:"provider"`
	ResolvedModel   string        `json:"resolved_model"`
	Dimensions      int           `json:"dimensions"`
	Vectors         [][]float64   `json:"vectors"`
	UsageReceipt    SignedReceipt `json:"usage_receipt"`
}

type UsageReceipt struct {
	ContractVersion     string     `json:"contract_version"`
	ReceiptID           string     `json:"receipt_id"`
	InvocationID        string     `json:"invocation_id"`
	RequestHash         string     `json:"request_hash"`
	Provider            string     `json:"provider"`
	ResolvedModel       string     `json:"resolved_model"`
	Usage               TokenUsage `json:"usage"`
	FinishClass         string     `json:"finish_class"`
	DurationMS          int64      `json:"duration_ms"`
	ProviderReceiptHash string     `json:"provider_receipt_hash,omitempty"`
	IssuedAt            time.Time  `json:"issued_at"`
}

type SignedReceipt struct {
	ContractVersion    string       `json:"contract_version"`
	Algorithm          string       `json:"algorithm"`
	SigningKeyID       string       `json:"signing_key_id"`
	SigningKeyRevision int64        `json:"signing_key_revision"`
	Receipt            UsageReceipt `json:"receipt"`
	Signature          string       `json:"signature"`
}
