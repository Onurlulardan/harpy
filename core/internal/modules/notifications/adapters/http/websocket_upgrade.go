package notificationhttp

import (
	"crypto/sha1"
	"encoding/base64"
	"net/http"
	"strings"

	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/domain"
)

const websocketGUID = "258EAFA5-E914-47DA-95CA-C5AB0DC85B11"

func upgradeSocket(w http.ResponseWriter, r *http.Request) (*socketConnection, error) {
	if err := validateSocketHandshake(r); err != nil {
		return nil, err
	}
	key := r.Header.Get("Sec-WebSocket-Key")
	hijacker, ok := w.(http.Hijacker)
	if !ok {
		return nil, errRealtimeUnavailable
	}
	connection, buffer, err := hijacker.Hijack()
	if err != nil {
		return nil, errRealtimeUnavailable
	}
	hash := sha1.Sum([]byte(key + websocketGUID))
	response := "HTTP/1.1 101 Switching Protocols\r\n" +
		"Upgrade: websocket\r\nConnection: Upgrade\r\n" +
		"Sec-WebSocket-Accept: " + base64.StdEncoding.EncodeToString(hash[:]) + "\r\n" +
		"Sec-WebSocket-Protocol: " + domain.RealtimeProtocol + "\r\n\r\n"
	if _, err = buffer.WriteString(response); err != nil || buffer.Flush() != nil {
		_ = connection.Close()
		return nil, errRealtimeUnavailable
	}
	return newSocketConnection(connection, buffer.Reader), nil
}

func validateSocketHandshake(r *http.Request) error {
	if r == nil {
		return errRealtimeUnavailable
	}
	decoded, keyErr := base64.StdEncoding.DecodeString(r.Header.Get("Sec-WebSocket-Key"))
	if r.Method != http.MethodGet || keyErr != nil || len(decoded) != 16 ||
		r.Header.Get("Sec-WebSocket-Version") != "13" ||
		!headerToken(r.Header.Values("Connection"), "upgrade") ||
		!headerToken(r.Header.Values("Upgrade"), "websocket") {
		return errRealtimeUnavailable
	}
	if _, err := domain.ParseOfferedProtocols(offeredProtocols(r)); err != nil {
		return errRealtimeUnavailable
	}
	return nil
}

func offeredProtocols(r *http.Request) string {
	if r == nil {
		return ""
	}
	return strings.Join(r.Header.Values("Sec-WebSocket-Protocol"), ",")
}

func headerToken(values []string, expected string) bool {
	for _, line := range values {
		for _, value := range strings.Split(line, ",") {
			if strings.EqualFold(strings.TrimSpace(value), expected) {
				return true
			}
		}
	}
	return false
}
