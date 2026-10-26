package devjournal

import (
	"context"
	"crypto/ed25519"
	"fmt"
	"io"
	"os"
	"path/filepath"

	recoveryv1 "github.com/Onurlulardan/harpy/contracts/recoverysigning/v1"
)

type RecoveryBundleResolver struct{ path string }

func NewRecoveryBundleResolver(path string) (*RecoveryBundleResolver, error) {
	if !filepath.IsAbs(path) {
		return nil, fmt.Errorf("development proposal trust path is invalid")
	}
	return &RecoveryBundleResolver{path: path}, nil
}

func (resolver *RecoveryBundleResolver) ResolveProposalKey(
	ctx context.Context,
	kid string,
) (ed25519.PublicKey, error) {
	if resolver == nil || ctx == nil || ctx.Err() != nil || kid == "" {
		return nil, fmt.Errorf("development proposal trust is unavailable")
	}
	file, err := os.Open(resolver.path)
	if err != nil {
		return nil, fmt.Errorf("open development proposal trust: %w", err)
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil || !info.Mode().IsRegular() ||
		info.Size() < 2 || info.Size() > recoveryv1.MaximumBundleBytes {
		return nil, fmt.Errorf("development proposal trust is invalid")
	}
	document, err := io.ReadAll(io.LimitReader(
		file, recoveryv1.MaximumBundleBytes+1,
	))
	if err != nil || int64(len(document)) != info.Size() {
		return nil, fmt.Errorf("read development proposal trust")
	}
	defer clear(document)
	bundle, err := recoveryv1.Parse(document)
	if err != nil {
		return nil, fmt.Errorf("parse development proposal trust: %w", err)
	}
	public, _, found := bundle.VerificationKey(
		recoveryv1.PurposeRecoveryControl, kid,
	)
	if !found {
		return nil, fmt.Errorf("development proposal signer is untrusted")
	}
	return append(ed25519.PublicKey(nil), public[:]...), nil
}

var _ ProposalKeyResolver = (*RecoveryBundleResolver)(nil)
