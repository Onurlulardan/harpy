package smtp

import (
	"net/mail"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"

	gomail "github.com/wneessen/go-mail"
)

var credentialPattern = regexp.MustCompile(`^[A-Za-z0-9_-]{22}\.[A-Za-z0-9_-]{43}$`)
var messageIDPattern = regexp.MustCompile(`^[a-z0-9][a-z0-9.-]{1,190}@[a-zA-Z0-9.-]{1,253}$`)

func (client *Client) buildMessage(value Message) (*gomail.Msg, error) {
	parsed, err := mail.ParseAddress(value.Recipient)
	if client == nil || err != nil || parsed.Name != "" ||
		parsed.Address != value.Recipient || !messageIDPattern.MatchString(value.MessageID) ||
		!validTenantSlug(value.TenantSlug) ||
		!credentialPattern.Match(value.Token) || value.Date.IsZero() {
		return nil, ErrContract
	}
	message := gomail.NewMsg()
	if err := message.FromFormat(client.settings.fromName, client.settings.from); err != nil {
		return nil, ErrContract
	}
	if err := message.To(value.Recipient); err != nil {
		return nil, ErrContract
	}
	message.Subject(client.settings.subject)
	message.SetMessageIDWithValue(value.MessageID)
	message.SetDateWithValue(value.Date.UTC())
	message.SetBulk()
	message.SetBodyWriter(
		gomail.TypeTextPlain,
		client.settings.text.writer(client.settings.baseURL, value.Token, value.TenantSlug),
	)
	message.AddAlternativeWriter(
		gomail.TypeTextHTML,
		client.settings.html.writer(client.settings.baseURL, value.Token, value.TenantSlug),
	)
	return message, nil
}

func validTenantSlug(value string) bool {
	if value == "" || len(value) > 100 || !utf8.ValidString(value) ||
		strings.TrimSpace(value) != value {
		return false
	}
	for _, character := range value {
		if unicode.IsControl(character) {
			return false
		}
	}
	return true
}
