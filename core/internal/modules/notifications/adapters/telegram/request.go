package telegram

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptrace"
	"regexp"
	"strings"

	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/domain"
	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/ports"
)

var (
	chatIDPattern = regexp.MustCompile(`^-?[1-9][0-9]{0,19}$`)
	tokenPattern  = regexp.MustCompile(`^[1-9][0-9]{4,15}:[A-Za-z0-9_-]{20,128}$`)
)

type telegramBody struct {
	ChatID    string `json:"chat_id"`
	Text      string `json:"text"`
	ParseMode string `json:"parse_mode,omitempty"`
}

func buildRequest(ctx context.Context, request ports.ProviderRequest,
	wrote *bool,
) (*http.Request, []byte, bool) {
	target, token := string(request.Target), string(request.Credential)
	metadata := request.Metadata
	if !request.DeliveryID.Valid() || !chatIDPattern.MatchString(target) ||
		!tokenPattern.MatchString(token) || request.Body == "" ||
		len(request.Body) > 4096 || metadata.TelegramAPIBase != "https://api.telegram.org" ||
		(request.TemplateKind != domain.TemplatePlainText &&
			request.TemplateKind != domain.TemplateSystemHTML) {
		return nil, nil, false
	}
	body := telegramBody{ChatID: target, Text: request.Body}
	if request.TemplateKind == domain.TemplateSystemHTML {
		body.ParseMode = "HTML"
	}
	encoded, err := json.Marshal(body)
	if err != nil {
		return nil, nil, false
	}
	endpoint := metadata.TelegramAPIBase + "/bot" + token + "/sendMessage"
	httpRequest, err := http.NewRequestWithContext(ctx, http.MethodPost,
		endpoint, bytes.NewReader(encoded))
	if err != nil {
		clear(encoded)
		return nil, nil, false
	}
	httpRequest.Header.Set("Content-Type", "application/json")
	trace := &httptrace.ClientTrace{WroteRequest: func(httptrace.WroteRequestInfo) {
		*wrote = true
	}}
	httpRequest = httpRequest.WithContext(httptrace.WithClientTrace(httpRequest.Context(), trace))
	return httpRequest, encoded, !strings.ContainsAny(target+token, "\r\n")
}
