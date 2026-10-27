package aiinferencev1

import (
	"math"

	canonicaljsonv1 "github.com/Onurlulardan/harpy/contracts/canonicaljson/v1"
)

func ValidateGeneration(request GenerationRequest, operation string) error {
	if operation != OperationGenerate && operation != OperationStream {
		return ErrInvalidRequest
	}
	if err := validateBinding(request.Binding, operation); err != nil ||
		len(request.Messages) < 1 || len(request.Messages) > MaximumMessages ||
		len(request.Tools) > MaximumTools ||
		request.Settings.MaximumOutputTokens < MinimumOutputTokens ||
		request.Settings.MaximumOutputTokens > MaximumOutputTokens ||
		request.Settings.TimeoutMS < MinimumTimeoutMS ||
		request.Settings.TimeoutMS > MaximumTimeoutMS {
		return ErrInvalidRequest
	}
	if value := request.Settings.Temperature; value != nil &&
		(math.IsNaN(*value) || math.IsInf(*value, 0) || *value < 0 || *value > 2) {
		return ErrInvalidRequest
	}
	total := 0
	for _, message := range request.Messages {
		total += len(message.Content)
		if total > MaximumInputTextBytes || !validMessage(message) {
			return ErrInvalidRequest
		}
	}
	if !validTools(request.Tools) {
		return ErrInvalidRequest
	}
	expected, err := CanonicalGenerationHash(request)
	if err != nil || !HashMatches(request.RequestHash, expected) {
		return ErrInvalidRequest
	}
	return nil
}

func validMessage(message Message) bool {
	if !validText(message.Content, MaximumMessageBytes, true) ||
		len(message.ToolProposals) > MaximumToolProposals {
		return false
	}
	switch message.Role {
	case "system", "user":
		return message.Content != "" && len(message.ToolProposals) == 0 && message.ToolResult == nil
	case "assistant":
		return (message.Content != "" || len(message.ToolProposals) > 0) &&
			message.ToolResult == nil && validProposals(message.ToolProposals)
	case "tool":
		return message.Content == "" && len(message.ToolProposals) == 0 &&
			message.ToolResult != nil && validToolResult(*message.ToolResult)
	default:
		return false
	}
}

func validTools(tools []ToolDefinition) bool {
	seen, total := map[string]bool{}, 0
	for _, tool := range tools {
		total += len(tool.InputSchema)
		if !validName(tool.Name) || seen[tool.Name] || !validText(tool.Description, 4096, true) ||
			len(tool.InputSchema) == 0 || len(tool.InputSchema) > MaximumToolSchemaBytes ||
			total > MaximumToolSchemaTotal {
			return false
		}
		seen[tool.Name] = true
		if _, err := canonicaljsonv1.CanonicalizeObject(tool.InputSchema,
			MaximumToolSchemaBytes); err != nil {
			return false
		}
	}
	return true
}

func validProposals(proposals []ToolProposal) bool {
	for _, proposal := range proposals {
		if !validProposalID(proposal.ProposalID) || !validName(proposal.Name) ||
			len(proposal.Arguments) == 0 || len(proposal.Arguments) > MaximumToolArgumentsBytes {
			return false
		}
		if _, err := canonicaljsonv1.CanonicalizeObject(proposal.Arguments,
			MaximumToolArgumentsBytes); err != nil {
			return false
		}
	}
	return true
}

func validToolResult(result ToolResult) bool {
	return validProposalID(result.ProposalID) && validName(result.Name) &&
		validText(result.Content, MaximumMessageBytes, false)
}
