package gcpkms

import (
	"encoding/json"
	"net/url"
	"path/filepath"
	"regexp"
	"strings"
)

const (
	externalAccountType = "external_account"
	jwtSubjectTokenType = "urn:ietf:params:oauth:token-type:jwt"
	audiencePrefix      = "//iam.googleapis.com/projects/"
)

var regionalSTSHost = regexp.MustCompile(`^sts\.[a-z0-9-]+\.rep\.googleapis\.com$`)

type workloadDocument struct {
	Type                           string                   `json:"type"`
	Audience                       string                   `json:"audience"`
	SubjectTokenType               string                   `json:"subject_token_type"`
	TokenURL                       string                   `json:"token_url"`
	ServiceAccountImpersonationURL string                   `json:"service_account_impersonation_url"`
	CredentialSource               workloadCredentialSource `json:"credential_source"`
	PrivateKey                     string                   `json:"private_key"`
	ClientSecret                   string                   `json:"client_secret"`
	RefreshToken                   string                   `json:"refresh_token"`
}

type workloadCredentialSource struct {
	File string `json:"file"`
}

func workloadCredentialDocumentAllowed(document []byte) bool {
	if len(document) == 0 {
		return true
	}
	var value workloadDocument
	if json.Unmarshal(document, &value) != nil || value.Type != externalAccountType ||
		value.SubjectTokenType != jwtSubjectTokenType ||
		!validWorkloadAudience(value.Audience) || !validSTSURL(value.TokenURL) ||
		!validImpersonationURL(value.ServiceAccountImpersonationURL) ||
		!filepath.IsAbs(value.CredentialSource.File) || value.PrivateKey != "" ||
		value.ClientSecret != "" || value.RefreshToken != "" {
		return false
	}
	return true
}

func validWorkloadAudience(value string) bool {
	if !strings.HasPrefix(value, audiencePrefix) || len(value) > 512 {
		return false
	}
	parts := strings.Split(value, "/")
	return len(parts) == 11 && parts[2] == "iam.googleapis.com" &&
		parts[3] == "projects" && parts[5] == "locations" && parts[6] == "global" &&
		parts[7] == "workloadIdentityPools" && parts[9] == "providers" &&
		parts[4] != "" && parts[8] != "" && parts[10] != ""
}

func validSTSURL(value string) bool {
	parsed, err := url.Parse(value)
	return err == nil && parsed.Scheme == "https" && parsed.User == nil &&
		parsed.RawQuery == "" && parsed.Fragment == "" && parsed.Path == "/v1/token" &&
		(parsed.Host == "sts.googleapis.com" || regionalSTSHost.MatchString(parsed.Host))
}

func validImpersonationURL(value string) bool {
	if value == "" {
		return true
	}
	parsed, err := url.Parse(value)
	return err == nil && parsed.Scheme == "https" && parsed.User == nil &&
		parsed.Host == "iamcredentials.googleapis.com" && parsed.RawQuery == "" &&
		parsed.Fragment == "" && strings.HasPrefix(
		parsed.Path, "/v1/projects/-/serviceAccounts/",
	) && strings.HasSuffix(parsed.Path, ":generateAccessToken")
}
