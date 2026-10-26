package smtp

import (
	"fmt"
	"io"
	"net/url"
	"strings"
)

const invitationPlaceholder = "{{INVITATION_URL}}"

type bodyTemplate struct {
	prefix []byte
	suffix []byte
}

func newBodyTemplate(value string) (bodyTemplate, error) {
	return newBodyTemplateFor(value, invitationPlaceholder)
}

func newBodyTemplateFor(value, placeholder string) (bodyTemplate, error) {
	if placeholder == "" || len(value) < len(placeholder) || len(value) > 32*1024 ||
		strings.Count(value, placeholder) != 1 {
		return bodyTemplate{}, ErrContract
	}
	parts := strings.SplitN(value, placeholder, 2)
	return bodyTemplate{
		prefix: []byte(parts[0]), suffix: []byte(parts[1]),
	}, nil
}

func (template bodyTemplate) writer(
	baseURL string, token []byte, tenantSlug string,
) func(io.Writer) (int64, error) {
	return template.writerRoute(baseURL, "/accept-invitation/", token, tenantSlug)
}

func (template bodyTemplate) writerRoute(
	baseURL, route string, token []byte, tenantSlug string,
) func(io.Writer) (int64, error) {
	return func(destination io.Writer) (int64, error) {
		parts := [][]byte{
			template.prefix, []byte(baseURL), []byte(route),
			token, []byte("?tenant="), []byte(url.QueryEscape(tenantSlug)),
			template.suffix,
		}
		var total int64
		for _, part := range parts {
			written, err := destination.Write(part)
			total += int64(written)
			if err != nil || written != len(part) {
				if err == nil {
					err = io.ErrShortWrite
				}
				return total, fmt.Errorf("write invitation template: %w", err)
			}
		}
		return total, nil
	}
}
