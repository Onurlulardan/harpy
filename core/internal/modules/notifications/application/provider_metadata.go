package application

import (
	"crypto/sha256"
	"encoding/binary"
	"net"
	"net/mail"
	"net/url"
	"strings"

	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/ports"
)

func validateProviderMetadata(kind ports.ProviderKind,
	metadata ports.ProviderMetadata,
) bool {
	switch kind {
	case ports.ProviderSMTP:
		if metadata.SMTPHost == "" || len(metadata.SMTPHost) > 253 ||
			len(metadata.SMTPUsername) > 320 ||
			strings.ContainsAny(metadata.SMTPUsername, "\r\n\x00") ||
			metadata.SMTPPort < 1 || metadata.SMTPPort > 65535 ||
			(metadata.SMTPSecurity != "starttls" && metadata.SMTPSecurity != "implicit_tls") ||
			metadata.TelegramAPIBase != "" || net.ParseIP(metadata.SMTPHost) != nil {
			return false
		}
		address, err := mail.ParseAddress(metadata.FromAddress)
		return err == nil && address.Address != "" && !strings.ContainsAny(metadata.SMTPHost, "/:@")
	case ports.ProviderTelegram:
		parsed, err := url.Parse(metadata.TelegramAPIBase)
		return err == nil && parsed.Scheme == "https" && parsed.Host == "api.telegram.org" &&
			parsed.Path == "" && parsed.RawQuery == "" && parsed.Fragment == "" &&
			parsed.User == nil && metadata.SMTPHost == "" && metadata.SMTPUsername == "" &&
			metadata.SMTPPort == 0 &&
			metadata.SMTPSecurity == "" && metadata.FromAddress == ""
	default:
		return false
	}
}

func providerMetadataDigest(kind ports.ProviderKind, metadata ports.ProviderMetadata) [32]byte {
	hash := sha256.New()
	parts := []string{string(kind), metadata.SMTPHost, metadata.SMTPUsername, metadata.SMTPSecurity,
		metadata.FromAddress, metadata.TelegramAPIBase}
	for _, value := range parts {
		var size [4]byte
		binary.BigEndian.PutUint32(size[:], uint32(len(value)))
		_, _ = hash.Write(size[:])
		_, _ = hash.Write([]byte(value))
	}
	var port [4]byte
	binary.BigEndian.PutUint32(port[:], uint32(metadata.SMTPPort))
	_, _ = hash.Write(port[:])
	var result [32]byte
	copy(result[:], hash.Sum(nil))
	return result
}
