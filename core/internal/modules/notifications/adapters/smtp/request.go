package smtp

import (
	"bytes"
	"net"
	"net/mail"
	"strings"

	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/domain"
	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/ports"
)

func parseRequest(request ports.ProviderRequest) (exchangeRequest, bool) {
	metadata := request.Metadata
	from, fromErr := mail.ParseAddress(metadata.FromAddress)
	targetText := string(request.Target)
	target, targetErr := mail.ParseAddress(targetText)
	username, password, credentialValid := smtpCredential(metadata, request.Credential)
	validKind := request.TemplateKind == domain.TemplatePlainText ||
		request.TemplateKind == domain.TemplateSystemHTML
	if !request.DeliveryID.Valid() || fromErr != nil || targetErr != nil ||
		from.Address == "" || target.Address != targetText || !credentialValid ||
		metadata.SMTPHost == "" ||
		len(metadata.SMTPHost) > 253 || strings.ContainsAny(metadata.SMTPHost, "/:@") ||
		net.ParseIP(metadata.SMTPHost) != nil || metadata.SMTPPort < 1 ||
		metadata.SMTPPort > 65535 || (metadata.SMTPSecurity != "starttls" &&
		metadata.SMTPSecurity != "implicit_tls") || !validKind ||
		request.Subject == "" || strings.ContainsAny(request.Subject, "\r\n") ||
		len(request.Subject) > 512 || request.Body == "" || len(request.Body) > 64*1024 {
		return exchangeRequest{}, false
	}
	message, err := buildMessage(request, from.Address, target.Address)
	if err != nil {
		return exchangeRequest{}, false
	}
	return exchangeRequest{host: metadata.SMTPHost, port: metadata.SMTPPort,
		security: metadata.SMTPSecurity, from: from.Address, target: target.Address,
		username: username, password: password, message: message}, true
}

func smtpCredential(metadata ports.ProviderMetadata, credential []byte) ([]byte, []byte, bool) {
	if metadata.SMTPUsername != "" {
		username := []byte(metadata.SMTPUsername)
		valid := len(username) <= 320 && !bytes.ContainsAny(username, "\r\n\x00") &&
			len(credential) > 0 && len(credential) <= 4096 &&
			!bytes.ContainsRune(credential, '\x00')
		return username, credential, valid
	}
	parts := bytes.Split(credential, []byte{0})
	valid := len(parts) == 2 && len(parts[0]) > 0 && len(parts[0]) <= 320 &&
		len(parts[1]) > 0 && len(parts[1]) <= 4096
	if !valid {
		return nil, nil, false
	}
	return parts[0], parts[1], true
}
