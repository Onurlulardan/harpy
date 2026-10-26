package passwordblocklist

import (
	"context"
	"crypto/sha256"
	"errors"
	"unicode/utf8"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/password"
	"golang.org/x/text/unicode/norm"
)

func (blocklist *Blocklist) Blocked(
	ctx context.Context,
	candidate []byte,
) (bool, error) {
	if ctx == nil || blocklist == nil || !utf8.Valid(candidate) ||
		!norm.NFC.IsNormal(candidate) {
		return false, ErrCandidateInvalid
	}
	count := utf8.RuneCount(candidate)
	if count < 1 || count > password.MaximumScalars || len(candidate) > password.MaximumScalars*4 {
		return false, ErrCandidateInvalid
	}
	if err := ctx.Err(); err != nil {
		return false, errors.Join(ErrUnavailable, err)
	}
	blocklist.mu.RLock()
	defer blocklist.mu.RUnlock()
	if !blocklist.open {
		return false, ErrUnavailable
	}
	digest := sha256.Sum256(candidate)
	found, err := blocklist.artifact.Contains(digest)
	clear(digest[:])
	if err != nil {
		return false, ErrUnavailable
	}
	if err := ctx.Err(); err != nil {
		return false, errors.Join(ErrUnavailable, err)
	}
	return found, nil
}
