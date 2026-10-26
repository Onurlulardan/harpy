package main

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

const (
	audienceEnvironment  = "GOOGLE_EXTERNAL_ACCOUNT_AUDIENCE"
	tokenTypeEnvironment = "GOOGLE_EXTERNAL_ACCOUNT_TOKEN_TYPE"
	jwtTokenType         = "urn:ietf:params:oauth:token-type:jwt"
	hostAudience         = "//iam.googleapis.com/projects/123456789012/locations/global/" +
		"workloadIdentityPools/harpy-host/providers/host-runtime"
)

type options struct {
	keyFile string
	keyID   string
	subject string
}

type executableResponse struct {
	Version        int    `json:"version"`
	Success        bool   `json:"success"`
	TokenType      string `json:"token_type"`
	ExpirationTime int64  `json:"expiration_time"`
	IDToken        string `json:"id_token"`
}

func main() {
	if err := run(os.Args[1:], os.Getenv, time.Now, rand.Reader, os.Stdout); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(
	arguments []string,
	getenv func(string) string,
	clock func() time.Time,
	random io.Reader,
	output io.Writer,
) error {
	config, err := parseOptions(arguments)
	if err != nil {
		return err
	}
	audience := getenv(audienceEnvironment)
	if audience != hostAudience || getenv(tokenTypeEnvironment) != jwtTokenType ||
		getenv("GOOGLE_EXTERNAL_ACCOUNT_INTERACTIVE") != "0" {
		return fmt.Errorf("external-account environment is invalid")
	}
	seed, err := os.ReadFile(config.keyFile)
	if err != nil || len(seed) != ed25519.SeedSize {
		clear(seed)
		return fmt.Errorf("workload signing seed is invalid")
	}
	privateKey := ed25519.NewKeyFromSeed(seed)
	clear(seed)
	defer clear(privateKey)
	now := clock().UTC().Truncate(time.Second)
	token, expiresAt, err := issueToken(config, audience, now, privateKey, random)
	if err != nil {
		return err
	}
	return json.NewEncoder(output).Encode(executableResponse{
		Version: 1, Success: true, TokenType: jwtTokenType,
		ExpirationTime: expiresAt.Unix(), IDToken: token,
	})
}

func parseOptions(arguments []string) (options, error) {
	flags := flag.NewFlagSet("workload-token", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	var value options
	flags.StringVar(&value.keyFile, "key", "", "workload signing seed")
	flags.StringVar(&value.keyID, "kid", "", "workload key ID")
	flags.StringVar(&value.subject, "subject", "", "workload subject")
	if err := flags.Parse(arguments); err != nil || flags.NArg() != 0 ||
		!filepath.IsAbs(value.keyFile) {
		return options{}, fmt.Errorf("workload token arguments are invalid")
	}
	expected := "harpy-host-" + value.keyID
	if value.keyID != "api" && value.keyID != "worker" ||
		value.subject != expected {
		return options{}, fmt.Errorf("workload identity is invalid")
	}
	return value, nil
}

func issueToken(
	config options,
	audience string,
	now time.Time,
	privateKey ed25519.PrivateKey,
	random io.Reader,
) (string, time.Time, error) {
	var identifier [16]byte
	if _, err := io.ReadFull(random, identifier[:]); err != nil {
		return "", time.Time{}, fmt.Errorf("generate workload token ID: %w", err)
	}
	expiresAt := now.Add(5 * time.Minute)
	header, err := json.Marshal(map[string]string{
		"alg": "EdDSA", "kid": config.keyID, "typ": "JWT",
	})
	if err != nil {
		return "", time.Time{}, err
	}
	payload, err := json.Marshal(map[string]any{
		"aud": audience, "exp": expiresAt.Unix(), "iat": now.Unix(),
		"iss": "harpy-host", "jti": base64.RawURLEncoding.EncodeToString(identifier[:]),
		"nbf": now.Add(-30 * time.Second).Unix(), "sub": config.subject,
	})
	if err != nil {
		return "", time.Time{}, err
	}
	wire := base64.RawURLEncoding.EncodeToString(header) + "." +
		base64.RawURLEncoding.EncodeToString(payload)
	signature := ed25519.Sign(privateKey, []byte(wire))
	return wire + "." + base64.RawURLEncoding.EncodeToString(signature), expiresAt, nil
}
