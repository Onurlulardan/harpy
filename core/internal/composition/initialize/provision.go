package initialize

import (
	"context"
	"encoding/hex"
	"io"
	"strings"

	"github.com/Onurlulardan/harpy/core/internal/platform/foundation"
	harpygoose "github.com/Onurlulardan/harpy/core/internal/platform/goose"
)

func runTenantProvision(stdout, stderr io.Writer, cfg foundation.Config, tenant string) int {
	tenantID, err := parseTenantID(tenant)
	if err != nil {
		_, _ = io.WriteString(stderr, foundation.Redact(err.Error(), cfg.Secrets())+"\n")
		return 2
	}
	err = harpygoose.ProvisionTenant(context.Background(), cfg, tenantID)
	if err != nil {
		_, _ = io.WriteString(stderr, foundation.Redact(err.Error(), cfg.Secrets())+"\n")
		return 5
	}
	_, _ = io.WriteString(stdout, "stage=goose status=ok database="+cfg.Database+"\n")
	return 0
}

func parseTenantID(value string) ([16]byte, error) {
	var tenantID [16]byte
	compact := strings.ReplaceAll(strings.ToLower(strings.TrimSpace(value)), "-", "")
	if len(compact) != 32 {
		return tenantID, errTenantID
	}
	decoded, err := hex.DecodeString(compact)
	if err != nil || len(decoded) != len(tenantID) {
		return tenantID, errTenantID
	}
	copy(tenantID[:], decoded)
	if tenantID == ([16]byte{}) {
		return tenantID, errTenantID
	}
	return tenantID, nil
}
