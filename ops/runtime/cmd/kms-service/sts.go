package main

import (
	"encoding/json"
	"mime"
	"net/http"
	"strings"
)

const (
	tokenGrantType         = "urn:ietf:params:oauth:grant-type:token-exchange"
	accessTokenType        = "urn:ietf:params:oauth:token-type:access_token"
	jwtTokenType           = "urn:ietf:params:oauth:token-type:jwt"
	cloudKMSScope          = "https://www.googleapis.com/auth/cloudkms"
	workloadAudiencePrefix = "//iam.googleapis.com/projects/"
	maxTokenRequestBytes   = 16 * 1024
)

type tokenResponse struct {
	AccessToken     string `json:"access_token"`
	IssuedTokenType string `json:"issued_token_type"`
	TokenType       string `json:"token_type"`
	ExpiresIn       int    `json:"expires_in"`
}

func tokenExchange(
	response http.ResponseWriter,
	request *http.Request,
	authorities ...*tokenAuthority,
) {
	response.Header().Set("Cache-Control", "no-store")
	response.Header().Set("Content-Type", "application/json")
	if request.Method != http.MethodPost || !validFormMediaType(request) {
		writeTokenError(response, http.StatusBadRequest)
		return
	}
	request.Body = http.MaxBytesReader(response, request.Body, maxTokenRequestBytes)
	if request.ParseForm() != nil || !validTokenForm(request) {
		writeTokenError(response, http.StatusBadRequest)
		return
	}
	authority := localTokenAuthority()
	if len(authorities) == 1 && authorities[0] != nil {
		authority = authorities[0]
	}
	accessToken := "harpy-local-kms"
	expiresIn := 3600
	if authority.profile == "host" {
		var err error
		accessToken, expiresIn, err = authority.exchange(request.PostForm)
		if err != nil {
			writeTokenError(response, http.StatusUnauthorized)
			return
		}
	}
	response.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(response).Encode(tokenResponse{
		AccessToken:     accessToken,
		IssuedTokenType: accessTokenType,
		TokenType:       "Bearer",
		ExpiresIn:       expiresIn,
	})
}

func validFormMediaType(request *http.Request) bool {
	mediaType, _, err := mime.ParseMediaType(request.Header.Get("Content-Type"))
	return err == nil && mediaType == "application/x-www-form-urlencoded"
}

func validTokenForm(request *http.Request) bool {
	form := request.PostForm
	return request.URL.RawQuery == "" &&
		form.Get("grant_type") == tokenGrantType &&
		form.Get("requested_token_type") == accessTokenType &&
		form.Get("subject_token_type") == jwtTokenType &&
		form.Get("subject_token") != "" && form.Get("scope") == cloudKMSScope &&
		strings.HasPrefix(form.Get("audience"), workloadAudiencePrefix)
}

func writeTokenError(response http.ResponseWriter, statusCode int) {
	response.WriteHeader(statusCode)
	_ = json.NewEncoder(response).Encode(map[string]string{"error": "invalid_request"})
}
