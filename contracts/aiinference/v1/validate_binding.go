package aiinferencev1

import (
	"errors"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/google/uuid"
)

var (
	ErrInvalidRequest = errors.New("AI inference request is invalid")
	labelPattern      = regexp.MustCompile(`^[a-z][a-z0-9._-]{0,63}$`)
	modelPattern      = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._:/-]{0,199}$`)
	hexPattern        = regexp.MustCompile(`^[0-9a-f]{64}$`)
	proposalPattern   = regexp.MustCompile(`^tool_[a-z0-9_]{1,58}$`)
)

func validateBinding(binding Binding, operation string) error {
	invocation, invocationErr := uuid.Parse(binding.InvocationID)
	receipt, receiptErr := uuid.Parse(binding.ReceiptID)
	if binding.ContractVersion != RequestContractVersion || invocationErr != nil ||
		receiptErr != nil || invocation.Version() != 7 || receipt.Version() != 7 ||
		binding.ConfigRevision < 1 || binding.AuthorizationVersion < 1 ||
		binding.Operation != operation || !hexPattern.MatchString(binding.RequestHash) ||
		len(binding.SignedContext) == 0 || len(binding.SignedContext) > 16<<10 ||
		len(binding.BaseURL) > 2048 || !modelPattern.MatchString(binding.Model) {
		return ErrInvalidRequest
	}
	if _, err := uuid.Parse(binding.TenantID); err != nil {
		return ErrInvalidRequest
	}
	if _, err := uuid.Parse(binding.ConfigID); err != nil {
		return ErrInvalidRequest
	}
	if !validProvider(binding.Provider) || !validCredential(binding) {
		return ErrInvalidRequest
	}
	return nil
}

func validProvider(provider string) bool {
	return provider == "openai" || provider == "anthropic" || provider == "gemini" ||
		provider == "ollama" || provider == "custom"
}

func validCredential(binding Binding) bool {
	present := len(binding.CredentialEnvelope) > 0
	if present != (binding.CredentialRevision > 0) {
		return false
	}
	if present {
		if _, err := CredentialEnvelopeDigest(binding.CredentialEnvelope); err != nil {
			return false
		}
	}
	if binding.Provider == "openai" || binding.Provider == "anthropic" ||
		binding.Provider == "gemini" {
		return present
	}
	return binding.CredentialRevision >= 0
}

func validText(value string, maximum int, empty bool) bool {
	return (empty || value != "") && len(value) <= maximum &&
		utf8.ValidString(value) && !strings.ContainsRune(value, 0)
}

func validName(value string) bool { return labelPattern.MatchString(value) }

func validProposalID(value string) bool { return proposalPattern.MatchString(value) }
