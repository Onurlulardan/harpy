package providerregistry

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/Onurlulardan/harpy/core/internal/modules/platformauthority/domain/artifact"
	"github.com/Onurlulardan/harpy/core/internal/modules/platformauthority/ports"
)

type Client struct {
	baseURL    string
	httpClient *http.Client
}

func New(baseURL string, httpClient *http.Client) (*Client, error) {
	parsed, err := url.Parse(baseURL)
	if err != nil || parsed.Scheme != "https" || parsed.Host == "" ||
		parsed.User != nil || parsed.RawQuery != "" || parsed.Fragment != "" ||
		strings.TrimRight(baseURL, "/") != baseURL || httpClient == nil {
		return nil, fmt.Errorf("SaaS authority registry configuration is invalid")
	}
	return &Client{baseURL: baseURL, httpClient: httpClient}, nil
}

func (client *Client) EnsureBinding(
	ctx context.Context,
	proof artifact.VerifiedBindingProof,
) error {
	_, err := client.exchange(
		ctx, http.MethodPost, "/v1/protected-authority-bindings", proof.Bytes,
	)
	return err
}

func (client *Client) EnsureProviderBinding(
	ctx context.Context,
	binding ports.ProviderBinding,
) error {
	document, err := canonicalProviderBinding(binding)
	if err != nil {
		return err
	}
	_, err = client.exchange(
		ctx, http.MethodPost, "/v1/protected-authority-bindings", document,
	)
	return err
}

func (client *Client) ExchangeProposal(
	ctx context.Context,
	proposal artifact.SignedProposal,
) ([]byte, error) {
	return client.exchange(
		ctx, http.MethodPost, "/v1/protected-authority-heads", proposal.Bytes,
	)
}

func (client *Client) LatestHead(
	context.Context,
	ports.HeadKey,
) (ports.ExternalHead, error) {
	return ports.ExternalHead{}, fmt.Errorf("latest SaaS head requires checkpoint inclusion evidence")
}
