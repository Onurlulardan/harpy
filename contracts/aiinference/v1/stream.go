package aiinferencev1

const (
	StreamEventDelta        = "delta"
	StreamEventToolProposal = "tool_proposal"
	StreamEventTerminal     = "terminal"

	TerminalSucceeded = "succeeded"
	TerminalFailed    = "failed"

	FailureProviderRejected    = "provider_rejected"
	FailureProviderTimeout     = "provider_timeout"
	FailureProviderUnavailable = "provider_unavailable"
	FailureProtocolInvalid     = "provider_protocol_invalid"
	FailureStreamTruncated     = "stream_truncated"
	FailureCallerCancelled     = "caller_cancelled"
	FailureBackpressure        = "backpressure"
)

type StreamEvent struct {
	ContractVersion string          `json:"contract_version"`
	Sequence        int64           `json:"sequence"`
	Type            string          `json:"type"`
	Delta           string          `json:"delta,omitempty"`
	ToolProposal    *ToolProposal   `json:"tool_proposal,omitempty"`
	Terminal        *StreamTerminal `json:"terminal,omitempty"`
}

type StreamTerminal struct {
	Status string            `json:"status"`
	Result *GenerationResult `json:"result,omitempty"`
	Error  *TerminalError    `json:"error,omitempty"`
}

type TerminalError struct {
	Class string `json:"class"`
}

func ValidStreamEventType(value string) bool {
	return value == StreamEventDelta || value == StreamEventToolProposal ||
		value == StreamEventTerminal
}

func ValidFailureClass(value string) bool {
	return value == FailureProviderRejected || value == FailureProviderTimeout ||
		value == FailureProviderUnavailable || value == FailureProtocolInvalid ||
		value == FailureStreamTruncated || value == FailureCallerCancelled ||
		value == FailureBackpressure
}

func ValidStreamTerminal(value StreamTerminal) bool {
	return (value.Status == TerminalSucceeded && value.Result != nil && value.Error == nil) ||
		(value.Status == TerminalFailed && value.Result == nil && value.Error != nil &&
			ValidFailureClass(value.Error.Class))
}
