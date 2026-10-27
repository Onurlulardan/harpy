package gcpkms

import (
	"runtime"
	"strings"
	"testing"
)

func TestWorkloadCredentialAllowsMetadataAndOIDCFederation(t *testing.T) {
	if !workloadCredentialDocumentAllowed(nil) {
		t.Fatal("ambient metadata identity was rejected")
	}
	path := "/var/run/secrets/harpy/oidc.jwt"
	if runtime.GOOS == "windows" {
		path = `C:\run\secrets\harpy\oidc.jwt`
	}
	document := `{
		"type":"external_account",
		"audience":"//iam.googleapis.com/projects/123456789/locations/global/workloadIdentityPools/harpy/providers/runtime",
		"subject_token_type":"urn:ietf:params:oauth:token-type:jwt",
		"token_url":"https://sts.googleapis.com/v1/token",
		"credential_source":{"file":"` + strings.ReplaceAll(path, `\`, `\\`) + `"}
	}`
	if !workloadCredentialDocumentAllowed([]byte(document)) {
		t.Fatal("bounded OIDC workload federation was rejected")
	}
}

func TestWorkloadCredentialRejectsLongLivedOrUntrustedCredential(t *testing.T) {
	values := []string{
		`{"type":"service_account","private_key":"secret"}`,
		`{"type":"authorized_user","refresh_token":"secret"}`,
		`{"type":"external_account","audience":"//iam.googleapis.com/projects/1/locations/global/workloadIdentityPools/p/providers/x","subject_token_type":"urn:ietf:params:oauth:token-type:jwt","token_url":"https://evil.invalid/v1/token","credential_source":{"file":"/run/token"}}`,
		`{"type":"external_account","audience":"//iam.googleapis.com/projects/1/locations/global/workloadIdentityPools/p/providers/x","subject_token_type":"urn:ietf:params:oauth:token-type:jwt","token_url":"https://sts.googleapis.com/v1/token","credential_source":{"file":"relative/token"}}`,
	}
	for _, value := range values {
		if workloadCredentialDocumentAllowed([]byte(value)) {
			t.Fatal("unsafe workload credential document was accepted")
		}
	}
}
