package telegram

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"io"
	"net/http"
	"strconv"

	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/ports"
)

type telegramResponse struct {
	OK     bool `json:"ok"`
	Result struct {
		MessageID int64 `json:"message_id"`
	} `json:"result"`
}

func (client *Client) Send(ctx context.Context,
	request ports.ProviderRequest,
) ports.ProviderResult {
	if client == nil || ctx == nil {
		return invalidResult(client, request)
	}
	wrote := false
	httpRequest, encoded, ok := buildRequest(ctx, request, &wrote)
	defer clear(encoded)
	if !ok {
		return invalidResult(client, request)
	}
	response, err := client.http.Do(httpRequest)
	if err != nil {
		return client.networkFailure(wrote)
	}
	defer response.Body.Close()
	completed := client.now().UTC()
	status := strconv.Itoa(response.StatusCode)
	if response.StatusCode == http.StatusTooManyRequests {
		return ports.ProviderResult{Outcome: ports.ProviderTransient,
			Code: "TELEGRAM_RATE_LIMITED", ResponseClass: "authenticated_429",
			StatusCode: status, CompletedAt: completed}
	}
	if response.StatusCode >= 500 {
		return ports.ProviderResult{Outcome: ports.ProviderUnknown,
			Code: "NOTIFICATION_OUTCOME_UNKNOWN", ResponseClass: "server_5xx",
			StatusCode: status, CompletedAt: completed}
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return ports.ProviderResult{Outcome: ports.ProviderPermanent,
			Code: "TELEGRAM_REJECTED", ResponseClass: "client_rejected",
			StatusCode: status, CompletedAt: completed}
	}
	var body telegramResponse
	decoder := json.NewDecoder(io.LimitReader(response.Body, 64*1024))
	if decoder.Decode(&body) != nil || !body.OK || body.Result.MessageID < 1 {
		return ports.ProviderResult{Outcome: ports.ProviderUnknown,
			Code: "NOTIFICATION_OUTCOME_UNKNOWN", ResponseClass: "invalid_response",
			StatusCode: status, CompletedAt: completed}
	}
	reference := sha256.Sum256([]byte(strconv.FormatInt(body.Result.MessageID, 10)))
	return ports.ProviderResult{Outcome: ports.ProviderAccepted,
		Code: "TELEGRAM_ACCEPTED", ResponseClass: "accepted",
		ProviderReferenceDigest: reference, StatusCode: status, CompletedAt: completed}
}

func (client *Client) networkFailure(wrote bool) ports.ProviderResult {
	outcome, code := ports.ProviderTransient, "TELEGRAM_CONNECT_FAILED"
	class := "pre_bytes"
	if wrote {
		outcome, code, class = ports.ProviderUnknown,
			"NOTIFICATION_OUTCOME_UNKNOWN", "after_bytes"
	}
	return ports.ProviderResult{Outcome: outcome, Code: code,
		ResponseClass: class, CompletedAt: client.now().UTC()}
}
