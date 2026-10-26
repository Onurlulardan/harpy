package smtp

import (
	"bytes"
	"fmt"
	"mime"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/domain"
	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/ports"
)

func stableMessageID(deliveryID string) string {
	return "<notification-" + deliveryID + "@harpy.invalid>"
}

func buildMessage(request ports.ProviderRequest, from, target string) ([]byte, error) {
	if request.Requested.IsZero() {
		return nil, fmt.Errorf("requested time is required")
	}
	contentType := "text/plain; charset=UTF-8"
	if request.TemplateKind == domain.TemplateSystemHTML {
		contentType = "text/html; charset=UTF-8"
	}
	var message bytes.Buffer
	headers := []string{
		"From: " + from,
		"To: " + target,
		"Date: " + request.Requested.UTC().Format(time.RFC1123Z),
		"Message-ID: " + stableMessageID(request.DeliveryID.String()),
		"Subject: " + mime.QEncoding.Encode("utf-8", request.Subject),
		"MIME-Version: 1.0",
		"Content-Type: " + contentType,
		"Content-Transfer-Encoding: 8bit",
	}
	for _, header := range headers {
		_, _ = message.WriteString(header + "\r\n")
	}
	_, _ = message.WriteString("\r\n")
	_, _ = message.WriteString(request.Body)
	return message.Bytes(), nil
}
