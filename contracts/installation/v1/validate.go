package installationv1

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/mail"
	"regexp"
	"strings"
	"unicode/utf8"

	canonicaljsonv1 "github.com/Onurlulardan/harpy/contracts/canonicaljson/v1"
)

var (
	digestPattern = regexp.MustCompile(`^[0-9a-f]{64}$`)
	slugPattern   = regexp.MustCompile(`^[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?$`)
)

func validWire(wire manifestWire) bool {
	return wire.ContractVersion == ContractVersion &&
		(wire.DeploymentMode == DeploymentOnPremise || wire.DeploymentMode == DeploymentSaaS) &&
		validText(wire.PrimaryTenant.Name, 128) &&
		slugPattern.MatchString(wire.PrimaryTenant.Slug) &&
		!strings.Contains(wire.PrimaryTenant.Slug, "--") &&
		validText(wire.FirstAdmin.DisplayName, 128) &&
		validEmail(wire.FirstAdmin.Email) && digestPattern.MatchString(wire.RequestSHA256)
}

func validText(value string, maximum int) bool {
	return value != "" && len(value) <= maximum && utf8.ValidString(value) &&
		strings.TrimSpace(value) == value
}

func validEmail(value string) bool {
	if !validText(value, 254) || value != strings.ToLower(value) {
		return false
	}
	parsed, err := mail.ParseAddress(value)
	return err == nil && parsed.Name == "" && parsed.Address == value &&
		strings.Contains(value, "@")
}

func requestDigest(wire manifestWire) string {
	request := struct {
		EnvironmentID  string         `json:"environment_id"`
		DeploymentMode DeploymentMode `json:"deployment_mode"`
		PrimaryTenant  TenantMetadata `json:"primary_tenant"`
		FirstAdmin     AdminMetadata  `json:"first_admin"`
	}{wire.EnvironmentID, wire.DeploymentMode, wire.PrimaryTenant, wire.FirstAdmin}
	raw, err := json.Marshal(request)
	if err != nil {
		return ""
	}
	canonical, err := canonicaljsonv1.CanonicalizeObject(raw, MaximumBytes)
	if err != nil {
		return ""
	}
	digest := sha256.Sum256(canonical)
	return hex.EncodeToString(digest[:])
}

func parseUUID(value string) ([16]byte, bool) {
	var result [16]byte
	if len(value) != 36 || value[8] != '-' || value[13] != '-' ||
		value[18] != '-' || value[23] != '-' || strings.ToLower(value) != value {
		return result, false
	}
	compact := strings.NewReplacer("-", "").Replace(value)
	decoded, err := hex.DecodeString(compact)
	if err != nil || len(decoded) != len(result) {
		return result, false
	}
	copy(result[:], decoded)
	return result, result != [16]byte{}
}
