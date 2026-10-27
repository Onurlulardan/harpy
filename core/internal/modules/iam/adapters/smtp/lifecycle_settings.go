package smtp

import (
	"net/mail"
	"strings"
	"time"

	lifecycledeliveryv1 "github.com/Onurlulardan/harpy/contracts/lifecycledelivery/v1"
)

const lifecyclePlaceholder = "{{LIFECYCLE_URL}}"

type LifecycleTemplateInput struct {
	Purpose lifecycledeliveryv1.Purpose
	Route   string
	Subject string
	Text    string
	HTML    string
}

type lifecycleTemplate struct {
	route, subject string
	text, html     bodyTemplate
}

type LifecycleSettings struct {
	host, username, password, from, fromName, baseURL string
	port                                              int
	timeout                                           time.Duration
	templates                                         map[lifecycledeliveryv1.Purpose]lifecycleTemplate
}

func NewLifecycleSettings(host string, port int, username, password, from,
	fromName, baseURL string, timeout time.Duration,
	inputs []LifecycleTemplateInput) (LifecycleSettings, error) {
	parsedFrom, fromErr := mail.ParseAddress(from)
	if !validHost(host) || port < 1 || port > 65535 || username == "" || password == "" ||
		fromErr != nil || parsedFrom.Name != "" || parsedFrom.Address != from ||
		fromName == "" || strings.ContainsAny(fromName, "\r\n") ||
		!strings.HasPrefix(baseURL, "http") || strings.HasSuffix(baseURL, "/") ||
		timeout < time.Second || timeout > 30*time.Second || len(inputs) != 4 {
		return LifecycleSettings{}, ErrContract
	}
	settings := LifecycleSettings{host: host, port: port, username: username,
		password: password, from: from, fromName: fromName, baseURL: baseURL,
		timeout: timeout, templates: make(map[lifecycledeliveryv1.Purpose]lifecycleTemplate, 4)}
	for _, input := range inputs {
		text, textErr := newBodyTemplateFor(input.Text, lifecyclePlaceholder)
		html, htmlErr := newBodyTemplateFor(input.HTML, lifecyclePlaceholder)
		if input.Purpose.Validate() != nil || !validLifecycleRoute(input.Route) ||
			input.Subject == "" || strings.ContainsAny(input.Subject, "\r\n") ||
			textErr != nil || htmlErr != nil || settings.templates[input.Purpose].route != "" {
			return LifecycleSettings{}, ErrContract
		}
		settings.templates[input.Purpose] = lifecycleTemplate{
			route: input.Route, subject: input.Subject, text: text, html: html,
		}
	}
	return settings, nil
}

func validLifecycleRoute(route string) bool {
	return route == "/verify-email/" || route == "/reset-password/" ||
		route == "/portal/setup/" || route == "/initial-admin/setup/"
}
