package providerregistry

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"cloud.google.com/go/auth/credentials"
	"cloud.google.com/go/auth/httptransport"
)

type workloadIdentityDocument struct {
	Type             string `json:"type"`
	PrivateKey       string `json:"private_key"`
	ClientSecret     string `json:"client_secret"`
	RefreshToken     string `json:"refresh_token"`
	CredentialSource struct {
		File string `json:"file"`
	} `json:"credential_source"`
}

func NewAuthenticatedHTTPClient(
	workloadIdentityPath string,
	configured bool,
) (*http.Client, error) {
	if configured {
		document, err := os.ReadFile(workloadIdentityPath)
		if err != nil {
			return nil, fmt.Errorf("read SaaS workload identity: %w", err)
		}
		defer clear(document)
		var identity workloadIdentityDocument
		if len(document) > 64*1024 || json.Unmarshal(document, &identity) != nil ||
			identity.Type != "external_account" ||
			!filepath.IsAbs(identity.CredentialSource.File) ||
			identity.PrivateKey != "" || identity.ClientSecret != "" ||
			identity.RefreshToken != "" {
			return nil, fmt.Errorf("SaaS authority workload identity is invalid")
		}
		if _, err = os.Stat(identity.CredentialSource.File); err != nil {
			return nil, fmt.Errorf("SaaS authority subject token is unavailable")
		}
	}
	options := &credentials.DetectOptions{
		Scopes: []string{"https://www.googleapis.com/auth/cloud-platform"},
	}
	if configured {
		options.CredentialsFile = workloadIdentityPath
	}
	identity, err := credentials.DetectDefault(options)
	if err != nil || identity == nil {
		return nil, fmt.Errorf("SaaS authority workload identity is unavailable")
	}
	client, err := httptransport.NewClient(
		&httptransport.Options{Credentials: identity},
	)
	if err != nil {
		return nil, fmt.Errorf("open SaaS authority HTTP transport: %w", err)
	}
	client.Timeout = 15 * time.Second
	return client, nil
}
