package application

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"net"
	"net/mail"
	"regexp"
	"strings"

	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/api"
)

var smtpSettingsKey = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._:-]{15,127}$`)

func normalizeSMTPReplace(value api.SMTPSettingsReplaceRequest) (
	api.SMTPSettingsReplaceRequest, bool,
) {
	value.Host = strings.ToLower(strings.TrimSpace(value.Host))
	value.Username = strings.TrimSpace(value.Username)
	value.From = strings.TrimSpace(value.From)
	valid := value.ExpectedSettingsRevision > 0 &&
		value.ExpectedSettingsRevision < int64(^uint64(0)>>1) &&
		value.ResultSettingsRevision == value.ExpectedSettingsRevision+1 &&
		smtpSettingsKey.MatchString(value.IdempotencyKey) &&
		value.RequestSHA256 != [32]byte{} && validSMTPHost(value.Host) &&
		value.Port > 0 && value.Port <= 65535 && value.Username != "" &&
		len(value.Username) <= 320 && !strings.ContainsAny(value.Username, "\r\n\x00") &&
		len(value.Password) <= 4096 && !bytes.ContainsRune(value.Password, '\x00') &&
		validSMTPAddress(value.From)
	return value, valid
}

func normalizeSMTPTest(value api.SMTPTestIntentRequest) (
	api.SMTPTestIntentRequest, bool,
) {
	value.To = strings.TrimSpace(value.To)
	parsed, err := mail.ParseAddress(value.To)
	if err == nil {
		value.To = parsed.Address
	}
	valid := err == nil && value.To != "" && len(value.To) <= 512 &&
		!strings.ContainsAny(value.To, "\r\n\x00") &&
		value.ExpectedSettingsRevision > 0 &&
		smtpSettingsKey.MatchString(value.IdempotencyKey) &&
		value.RequestSHA256 != [32]byte{} && !value.RequestedAt.IsZero()
	return value, valid
}

func validSMTPHost(value string) bool {
	if value == "" || len(value) > 253 || net.ParseIP(value) != nil ||
		strings.ContainsAny(value, "/:@\r\n\x00") {
		return false
	}
	for _, label := range strings.Split(value, ".") {
		if label == "" || len(label) > 63 || label[0] == '-' ||
			label[len(label)-1] == '-' {
			return false
		}
		for _, character := range label {
			if character != '-' && (character < 'a' || character > 'z') &&
				(character < '0' || character > '9') {
				return false
			}
		}
	}
	return true
}

func validSMTPAddress(value string) bool {
	parsed, err := mail.ParseAddress(value)
	return err == nil && parsed.Address != "" && len(value) <= 512 &&
		!strings.ContainsAny(value, "\r\n\x00")
}

func smtpReplaceDigest(value api.SMTPSettingsReplaceRequest) [32]byte {
	document := append([]byte("harpy/notifications/smtp-settings/replace/v1"),
		value.RequestSHA256[:]...)
	document = appendSMTPField(document, []byte(value.Host))
	document = binary.BigEndian.AppendUint32(document, uint32(value.Port))
	document = appendSMTPField(document, []byte(value.Username))
	document = appendSMTPField(document, []byte(value.From))
	secretDigest := sha256.Sum256(value.Password)
	document = append(document, secretDigest[:]...)
	document = binary.BigEndian.AppendUint64(document,
		uint64(value.ExpectedSettingsRevision))
	document = binary.BigEndian.AppendUint64(document,
		uint64(value.ResultSettingsRevision))
	result := sha256.Sum256(document)
	clear(document)
	return result
}

func smtpTestDigest(value api.SMTPTestIntentRequest) [32]byte {
	document := append([]byte("harpy/notifications/smtp-settings/test/v1"),
		value.RequestSHA256[:]...)
	document = appendSMTPField(document, []byte(value.To))
	document = binary.BigEndian.AppendUint64(document,
		uint64(value.ExpectedSettingsRevision))
	result := sha256.Sum256(document)
	clear(document)
	return result
}

func appendSMTPField(document, value []byte) []byte {
	document = binary.BigEndian.AppendUint32(document, uint32(len(value)))
	return append(document, value...)
}
