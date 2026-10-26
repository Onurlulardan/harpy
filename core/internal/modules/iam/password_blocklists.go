package iam

import (
	"errors"
	"fmt"
	"sync"

	blocklistv1 "github.com/Onurlulardan/harpy/contracts/passwordblocklist/v1"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/adapters/passwordblocklist"
	"github.com/Onurlulardan/harpy/core/internal/platform/config"
)

type PasswordBlocklists struct {
	mu          sync.Mutex
	common      *passwordblocklist.Blocklist
	compromised *passwordblocklist.Blocklist
	closed      bool
}

func OpenPasswordBlocklists(
	configured config.IAMPasswordBlocklists,
) (*PasswordBlocklists, error) {
	if configured.Validate() != nil {
		return nil, fmt.Errorf("IAM password blocklist config is invalid")
	}
	commonConfig, _ := configured.Common()
	common, err := openPasswordBlocklist(commonConfig, blocklistv1.KindCommon)
	if err != nil {
		return nil, fmt.Errorf("open common password blocklist: %w", err)
	}
	compromisedConfig, _ := configured.Compromised()
	compromised, err := openPasswordBlocklist(
		compromisedConfig, blocklistv1.KindCompromised,
	)
	if err != nil {
		_ = common.Close()
		return nil, fmt.Errorf("open compromised password blocklist: %w", err)
	}
	return &PasswordBlocklists{common: common, compromised: compromised}, nil
}

func openPasswordBlocklist(
	configured config.IAMPasswordBlocklistArtifact,
	kind blocklistv1.Kind,
) (*passwordblocklist.Blocklist, error) {
	path, err := configured.Path()
	if err != nil {
		return nil, err
	}
	digest, err := configured.SHA256()
	if err != nil {
		return nil, err
	}
	return passwordblocklist.Open(path.Value(), digest, kind)
}

func (blocklists *PasswordBlocklists) Common() IdentityPasswordRejector {
	if blocklists == nil {
		return nil
	}
	return blocklists.common
}

func (blocklists *PasswordBlocklists) Compromised() IdentityPasswordRejector {
	if blocklists == nil {
		return nil
	}
	return blocklists.compromised
}

func (blocklists *PasswordBlocklists) Close() error {
	if blocklists == nil {
		return nil
	}
	blocklists.mu.Lock()
	defer blocklists.mu.Unlock()
	if blocklists.closed {
		return nil
	}
	blocklists.closed = true
	return errors.Join(blocklists.common.Close(), blocklists.compromised.Close())
}
