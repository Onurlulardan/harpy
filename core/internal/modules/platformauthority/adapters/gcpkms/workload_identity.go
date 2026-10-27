package gcpkms

import (
	"encoding/json"
	"net/url"
	"path/filepath"
	"regexp"
	"strings"
)

const workloadType = "external_account"
const workloadJWT = "urn:ietf:params:oauth:token-type:jwt"

var workloadSTSHost = regexp.MustCompile(
	`^sts(?:\.[a-z0-9-]+\.rep)?\.googleapis\.com$`,
)

type workloadDocument struct {
	Type             string `json:"type"`
	Audience         string `json:"audience"`
	SubjectTokenType string `json:"subject_token_type"`
	TokenURL         string `json:"token_url"`
	ImpersonationURL string `json:"service_account_impersonation_url"`
	CredentialSource struct {
		File string `json:"file"`
	} `json:"credential_source"`
	PrivateKey, ClientSecret, RefreshToken string
}

func workloadCredentialDocumentAllowed(document []byte) bool {
	if len(document) == 0 {
		return true
	}
	var value workloadDocument
	if json.Unmarshal(document, &value) != nil || value.Type != workloadType ||
		value.SubjectTokenType != workloadJWT || !validAudience(value.Audience) ||
		!validTokenURL(value.TokenURL) || !validImpersonationURL(value.ImpersonationURL) ||
		!filepath.IsAbs(value.CredentialSource.File) || value.PrivateKey != "" ||
		value.ClientSecret != "" || value.RefreshToken != "" {
		return false
	}
	return true
}

func validAudience(value string) bool {
	parts := strings.Split(value, "/")
	return strings.HasPrefix(value, "//iam.googleapis.com/projects/") &&
		len(value) <= 512 && len(parts) == 11 &&
		parts[2] == "iam.googleapis.com" && parts[3] == "projects" &&
		parts[4] != "" && parts[5] == "locations" && parts[6] == "global" &&
		parts[7] == "workloadIdentityPools" && parts[8] != "" &&
		parts[9] == "providers" && parts[10] != ""
}

func validTokenURL(value string) bool {
	parsed, err := url.Parse(value)
	return err == nil && parsed.Scheme == "https" && parsed.User == nil &&
		parsed.RawQuery == "" && parsed.Fragment == "" && parsed.Path == "/v1/token" &&
		workloadSTSHost.MatchString(parsed.Host)
}

func validImpersonationURL(value string) bool {
	if value == "" {
		return true
	}
	parsed, err := url.Parse(value)
	return err == nil && parsed.Scheme == "https" && parsed.User == nil &&
		parsed.Host == "iamcredentials.googleapis.com" && parsed.RawQuery == "" &&
		parsed.Fragment == "" && strings.HasPrefix(parsed.Path,
		"/v1/projects/-/serviceAccounts/") &&
		strings.HasSuffix(parsed.Path, ":generateAccessToken")
}
