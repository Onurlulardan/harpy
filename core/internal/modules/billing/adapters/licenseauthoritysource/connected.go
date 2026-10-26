package licenseauthoritysource

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	licenseauthority "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/licenseauthority"
)

const issuePath = "/api/v2/license-snapshots/issue"

type Connected struct {
	endpoint   string
	credential string
	client     *http.Client
}

func NewConnected(
	baseURL string,
	credentialDocument []byte,
	transport http.RoundTripper,
	timeout time.Duration,
) (*Connected, error) {
	parsed, err := url.ParseRequestURI(baseURL)
	if err != nil || parsed.Host == "" || parsed.User != nil || parsed.RawQuery != "" ||
		parsed.Fragment != "" || strings.TrimSpace(baseURL) != baseURL || timeout <= 0 {
		return nil, ErrInvalidConfiguration
	}
	if parsed.Scheme != "https" &&
		(parsed.Scheme != "http" || !loopbackHost(parsed.Hostname())) {
		return nil, ErrInvalidConfiguration
	}
	credential, err := parseCredential(credentialDocument)
	if err != nil {
		return nil, err
	}
	endpoint, err := url.JoinPath(baseURL, issuePath)
	if err != nil {
		return nil, ErrInvalidConfiguration
	}
	if transport == nil {
		transport = http.DefaultTransport
	}
	return &Connected{
		endpoint: endpoint, credential: credential,
		client: &http.Client{
			Transport: transport, Timeout: timeout,
			CheckRedirect: func(*http.Request, []*http.Request) error {
				return http.ErrUseLastResponse
			},
		},
	}, nil
}

func loopbackHost(host string) bool {
	if strings.EqualFold(host, "localhost") {
		return true
	}
	address := net.ParseIP(host)
	return address != nil && address.IsLoopback()
}

func (source *Connected) IssueSnapshot(
	ctx context.Context, request licenseauthority.IssueRequest,
) ([]byte, error) {
	body, err := marshalIssueRequest(request)
	if err != nil {
		return nil, fmt.Errorf("encode license issue request: %w", err)
	}
	httpRequest, err := http.NewRequestWithContext(
		ctx, http.MethodPost, source.endpoint, bytes.NewReader(body),
	)
	if err != nil {
		return nil, fmt.Errorf("build license issue request: %w", err)
	}
	httpRequest.Header.Set("Authorization", "Bearer "+source.credential)
	httpRequest.Header.Set("Content-Type", "application/json")
	httpRequest.Header.Set("Accept", "application/json")
	response, err := source.client.Do(httpRequest)
	if err != nil {
		return nil, fmt.Errorf("send license issue request: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK || !isJSON(response.Header.Get("Content-Type")) {
		_, _ = io.Copy(io.Discard, io.LimitReader(response.Body, 4096))
		return nil, fmt.Errorf("license issuer returned status %d", response.StatusCode)
	}
	document, err := io.ReadAll(io.LimitReader(
		response.Body, licenseauthority.MaximumDocumentBytes+1,
	))
	if err != nil || len(document) == 0 || len(document) > licenseauthority.MaximumDocumentBytes {
		return nil, fmt.Errorf("license issuer response is invalid")
	}
	return document, nil
}

func isJSON(contentType string) bool {
	mediaType, _, err := mime.ParseMediaType(contentType)
	return err == nil && mediaType == "application/json"
}

var _ licenseauthority.Source = (*Connected)(nil)
