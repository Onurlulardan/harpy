package password

import (
	"context"
	"errors"
	"strings"
	"unicode/utf8"

	passwordpolicy "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/passwordpolicy"
	"golang.org/x/text/cases"
	"golang.org/x/text/unicode/norm"
)

const (
	MinimumScalars = passwordpolicy.MinimumScalars
	MaximumScalars = passwordpolicy.MaximumScalars
)

type Rejector interface {
	Blocked(context.Context, []byte) (bool, error)
}

type Policy struct {
	common      Rejector
	compromised Rejector
	valid       bool
}

type Secret struct {
	value     []byte
	destroyed bool
}

func NewPolicy(common Rejector, compromised Rejector) (Policy, error) {
	if common == nil || compromised == nil {
		return Policy{}, ErrPasswordPolicyUnavailable
	}
	return Policy{common: common, compromised: compromised, valid: true}, nil
}

func (policy Policy) NewSecret(
	ctx context.Context,
	candidate []byte,
	contextTerms []string,
) (*Secret, error) {
	if ctx == nil || !policy.valid {
		return nil, ErrPasswordPolicyUnavailable
	}
	normalized, err := normalize(candidate, MinimumScalars)
	if err != nil {
		return nil, err
	}
	for _, rejector := range []Rejector{policy.common, policy.compromised} {
		rejected, err := rejector.Blocked(ctx, normalized)
		if err != nil {
			wipe(normalized)
			return nil, errors.Join(ErrPasswordPolicyUnavailable, err)
		}
		if rejected {
			wipe(normalized)
			return nil, ErrPasswordRejected
		}
	}
	if contextRejected(normalized, contextTerms) {
		wipe(normalized)
		return nil, ErrPasswordRejected
	}
	return &Secret{value: normalized}, nil
}

func contextRejected(candidate []byte, terms []string) bool {
	foldedCandidate := cases.Fold().String(string(candidate))
	for _, term := range terms {
		if !utf8.ValidString(term) {
			continue
		}
		foldedTerm := cases.Fold().String(norm.NFC.String(strings.TrimSpace(term)))
		if utf8.RuneCountInString(foldedTerm) >= 4 &&
			strings.Contains(foldedCandidate, foldedTerm) {
			return true
		}
	}
	return false
}

func NewVerificationSecret(candidate []byte) (*Secret, error) {
	normalized, err := normalize(candidate, 1)
	if err != nil {
		return nil, err
	}
	return &Secret{value: normalized}, nil
}

func normalize(candidate []byte, minimum int) ([]byte, error) {
	if !utf8.Valid(candidate) {
		return nil, ErrPasswordInvalid
	}
	normalized := norm.NFC.Bytes(candidate)
	count := utf8.RuneCount(normalized)
	if count < minimum || count > MaximumScalars {
		wipe(normalized)
		return nil, ErrPasswordInvalid
	}
	return normalized, nil
}

func (secret *Secret) Destroy() {
	if secret == nil || secret.destroyed {
		return
	}
	wipe(secret.value)
	secret.value = nil
	secret.destroyed = true
}

func (secret *Secret) usable() bool {
	return secret != nil && !secret.destroyed && len(secret.value) > 0
}

func wipe(value []byte) {
	for index := range value {
		value[index] = 0
	}
}
