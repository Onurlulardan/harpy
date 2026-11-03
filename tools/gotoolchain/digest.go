package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"
)

const bookwormDigestAccept = "application/vnd.oci.image.index.v1+json, " +
	"application/vnd.docker.distribution.manifest.list.v2+json"

var imageDigestPattern = regexp.MustCompile(`^sha256:[0-9a-f]{64}$`)

type registryClient struct {
	tokenURL    string
	manifestURL string
	client      *http.Client
}

func bookwormDigest(ctx context.Context, version goRelease) (string, error) {
	client := registryClient{
		tokenURL: "https://auth.docker.io/token?service=registry.docker.io" +
			"&scope=repository:library/golang:pull",
		manifestURL: "https://registry-1.docker.io/v2/library/golang/manifests/",
		client:      &http.Client{Timeout: 20 * time.Second},
	}
	return client.bookwormDigest(ctx, version)
}

func (registry registryClient) bookwormDigest(
	ctx context.Context,
	version goRelease,
) (string, error) {
	if registry.client == nil {
		return "", fmt.Errorf("registry client is required")
	}
	token, err := registry.token(ctx)
	if err != nil {
		return "", err
	}
	tag := version.String() + "-bookworm"
	request, err := http.NewRequestWithContext(
		ctx, http.MethodGet, registry.manifestURL+tag, nil,
	)
	if err != nil {
		return "", fmt.Errorf("build manifest request: %w", err)
	}
	request.Header.Set("Authorization", "Bearer "+token)
	request.Header.Set("Accept", bookwormDigestAccept)
	response, err := registry.client.Do(request)
	if err != nil {
		return "", fmt.Errorf("fetch golang image manifest: %w", err)
	}
	defer response.Body.Close()
	_, _ = io.Copy(io.Discard, response.Body)
	if response.StatusCode == http.StatusNotFound {
		return "", fmt.Errorf("golang image %s was not found", tag)
	}
	if response.StatusCode != http.StatusOK {
		return "", fmt.Errorf("golang image manifest returned %s", response.Status)
	}
	digest := response.Header.Get("Docker-Content-Digest")
	if !imageDigestPattern.MatchString(digest) {
		return "", fmt.Errorf("golang image manifest did not return an exact sha256 digest")
	}
	return digest, nil
}

func (registry registryClient) token(ctx context.Context) (string, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, registry.tokenURL, nil)
	if err != nil {
		return "", fmt.Errorf("build registry token request: %w", err)
	}
	response, err := registry.client.Do(request)
	if err != nil {
		return "", fmt.Errorf("fetch registry token: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		_, _ = io.Copy(io.Discard, response.Body)
		return "", fmt.Errorf("registry token returned %s", response.Status)
	}
	var payload struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		return "", fmt.Errorf("decode registry token: %w", err)
	}
	if payload.Token == "" {
		return "", fmt.Errorf("registry token is empty")
	}
	return payload.Token, nil
}

func validImageDigest(digest string) bool {
	return imageDigestPattern.MatchString(strings.TrimSpace(digest))
}
