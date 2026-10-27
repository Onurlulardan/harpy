package smtp

import (
	"crypto/tls"
	"net/mail"
	"net/url"
	"strings"
	"time"

	gomail "github.com/wneessen/go-mail"
)

type Settings struct {
	host     string
	port     int
	username string
	password string
	from     string
	fromName string
	baseURL  string
	subject  string
	text     bodyTemplate
	html     bodyTemplate
	timeout  time.Duration
}

func NewSettings(
	host string,
	port int,
	username string,
	password string,
	from string,
	fromName string,
	baseURL string,
	subject string,
	textBody string,
	htmlBody string,
	timeout time.Duration,
) (Settings, error) {
	textTemplate, textErr := newBodyTemplate(textBody)
	htmlTemplate, htmlErr := newBodyTemplate(htmlBody)
	parsedFrom, fromErr := mail.ParseAddress(from)
	if !validHost(host) || port < 1 || port > 65535 || username == "" ||
		password == "" || fromErr != nil || parsedFrom.Name != "" ||
		parsedFrom.Address != from || fromName == "" || strings.ContainsAny(fromName, "\r\n") ||
		!validBaseURL(baseURL) ||
		subject == "" || strings.ContainsAny(subject, "\r\n") || textErr != nil ||
		htmlErr != nil || timeout < time.Second || timeout > 30*time.Second {
		return Settings{}, ErrContract
	}
	return Settings{
		host: host, port: port, username: username, password: password,
		from: from, fromName: fromName, baseURL: baseURL, subject: subject,
		text: textTemplate, html: htmlTemplate, timeout: timeout,
	}, nil
}

func validBaseURL(value string) bool {
	parsed, err := url.Parse(value)
	return err == nil && parsed != nil &&
		(parsed.Scheme == "http" || parsed.Scheme == "https") &&
		parsed.Host != "" && parsed.User == nil && parsed.Path == "" &&
		parsed.RawQuery == "" && parsed.Fragment == "" && parsed.String() == value
}

func (settings Settings) clientOptions() []gomail.Option {
	return []gomail.Option{
		gomail.WithPort(settings.port), gomail.WithTimeout(settings.timeout),
		gomail.WithTLSPolicy(gomail.TLSMandatory),
		gomail.WithTLSConfig(&tls.Config{
			MinVersion: tls.VersionTLS12, ServerName: settings.host,
		}),
		gomail.WithSMTPAuth(gomail.SMTPAuthAutoDiscover),
		gomail.WithUsername(settings.username), gomail.WithPassword(settings.password),
	}
}

func validHost(value string) bool {
	if value == "" || len(value) > 253 || strings.Contains(value, "..") ||
		strings.ContainsAny(value, ":/[]@ \t\r\n") {
		return false
	}
	return true
}
