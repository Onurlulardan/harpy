package domain

import (
	"errors"
	"html"
	"sort"
	"strings"
)

var ErrTemplateInvalid = errors.New("notification template is invalid")

type TemplateKind string

const (
	TemplateSystemHTML TemplateKind = "system_html"
	TemplatePlainText  TemplateKind = "plain_text"
)

type PublishedTemplate struct {
	ID, VersionID               ID
	NotificationType            string
	Kind                        TemplateKind
	Locale                      string
	SubjectPattern, BodyPattern string
	Variables                   []string
	MandatoryEmail              bool
	Revision                    int64
}

type RenderedTemplate struct {
	Locale, Subject, Body string
}

func (value PublishedTemplate) Valid() bool {
	if !value.ID.Valid() || !value.VersionID.Valid() || value.Revision < 1 ||
		value.NotificationType == "" || value.Locale == "" ||
		value.SubjectPattern == "" || value.BodyPattern == "" ||
		(value.Kind != TemplateSystemHTML && value.Kind != TemplatePlainText) ||
		len(value.SubjectPattern) > 4*1024 || len(value.BodyPattern) > 64*1024 ||
		len(value.Variables) > 64 {
		return false
	}
	seen := make(map[string]struct{}, len(value.Variables))
	for _, variable := range value.Variables {
		if !validVariable(variable) {
			return false
		}
		if _, exists := seen[variable]; exists {
			return false
		}
		seen[variable] = struct{}{}
	}
	return templatePatternsValid(value.SubjectPattern, value.BodyPattern, seen)
}

func (value PublishedTemplate) Render(data map[string]string) (RenderedTemplate, error) {
	if !value.Valid() || len(data) != len(value.Variables) {
		return RenderedTemplate{}, ErrTemplateInvalid
	}
	keys := append([]string(nil), value.Variables...)
	sort.Strings(keys)
	total := 0
	subject, body := value.SubjectPattern, value.BodyPattern
	for _, key := range keys {
		input, exists := data[key]
		total += len(key) + len(input)
		if !exists || total > 16*1024 {
			return RenderedTemplate{}, ErrTemplateInvalid
		}
		replacement := input
		if value.Kind == TemplateSystemHTML {
			replacement = html.EscapeString(input)
		} else if strings.ContainsAny(input, "<>") {
			return RenderedTemplate{}, ErrTemplateInvalid
		}
		token := "{{" + key + "}}"
		if !strings.Contains(subject, token) && !strings.Contains(body, token) {
			return RenderedTemplate{}, ErrTemplateInvalid
		}
		subject = strings.ReplaceAll(subject, token, replacement)
		body = strings.ReplaceAll(body, token, replacement)
	}
	if strings.Contains(subject, "{{") || strings.Contains(body, "{{") ||
		len(subject) > 512 || len(body) > 64*1024 {
		return RenderedTemplate{}, ErrTemplateInvalid
	}
	return RenderedTemplate{Locale: value.Locale, Subject: subject, Body: body}, nil
}

func validVariable(value string) bool {
	if value == "" || len(value) > 64 {
		return false
	}
	for index, character := range value {
		if !(character >= 'a' && character <= 'z') && character != '_' &&
			!(index > 0 && character >= '0' && character <= '9') {
			return false
		}
	}
	return true
}
