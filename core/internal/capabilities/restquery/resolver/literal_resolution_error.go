package resolver

import (
	"fmt"
	"io"

	restqueryv1 "github.com/Onurlulardan/harpy/contracts/restquery/v1"
)

const resolutionRedaction = "[REDACTED]"

// LiteralResolutionError exposes only a stable issue code, never rejected data.
type LiteralResolutionError struct {
	code  restqueryv1.IssueCode
	valid bool
}

func newLiteralResolutionError(code restqueryv1.IssueCode) LiteralResolutionError {
	return LiteralResolutionError{code: code, valid: code.Valid()}
}

func (resolutionError LiteralResolutionError) Code() (restqueryv1.IssueCode, error) {
	if !resolutionError.valid || !resolutionError.code.Valid() {
		return "", fmt.Errorf("REST Query literal resolution error is invalid")
	}
	return resolutionError.code, nil
}

func (LiteralResolutionError) Error() string    { return resolutionRedaction }
func (LiteralResolutionError) String() string   { return resolutionRedaction }
func (LiteralResolutionError) GoString() string { return resolutionRedaction }
func (LiteralResolutionError) MarshalJSON() ([]byte, error) {
	return nil, fmt.Errorf("REST Query literal resolution errors cannot be marshaled directly")
}
func (LiteralResolutionError) Format(state fmt.State, _ rune) {
	_, _ = io.WriteString(state, resolutionRedaction)
}

func literalResolutionFailure(code restqueryv1.IssueCode) error {
	return newLiteralResolutionError(code)
}
