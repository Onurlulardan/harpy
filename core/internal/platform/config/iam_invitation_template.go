package config

import (
	"encoding/json"
	"io"
	"os"
	"strings"
)

const invitationURLPlaceholder = "{{INVITATION_URL}}"

type IAMInvitationTemplate struct {
	subject string
	text    string
	html    string
}

type invitationTemplateDocument struct {
	Version int    `json:"version"`
	Subject string `json:"subject"`
	Text    string `json:"text"`
	HTML    string `json:"html"`
}

func loadIAMInvitationTemplate(
	values environmentValues,
) (IAMInvitationTemplate, error) {
	path, err := loadAbsolutePath(values, "HARPY_IAM_INVITATION_TEMPLATE_FILE")
	if err != nil {
		return IAMInvitationTemplate{}, err
	}
	file, err := os.Open(path.Value())
	if err != nil {
		return IAMInvitationTemplate{}, configError(
			"HARPY_IAM_INVITATION_TEMPLATE_FILE", "cannot be opened",
		)
	}
	defer file.Close()
	decoder := json.NewDecoder(io.LimitReader(file, 64*1024+1))
	decoder.DisallowUnknownFields()
	var document invitationTemplateDocument
	if err := decoder.Decode(&document); err != nil || decoder.Decode(&struct{}{}) != io.EOF {
		return IAMInvitationTemplate{}, configError(
			"HARPY_IAM_INVITATION_TEMPLATE_FILE", "must satisfy iam-invitation-template.v1",
		)
	}
	if document.Version != 1 || len(document.Subject) < 1 ||
		len(document.Subject) > 200 || strings.ContainsAny(document.Subject, "\r\n") ||
		!validInvitationBody(document.Text) || !validInvitationBody(document.HTML) {
		return IAMInvitationTemplate{}, configError(
			"HARPY_IAM_INVITATION_TEMPLATE_FILE", "must satisfy iam-invitation-template.v1",
		)
	}
	return IAMInvitationTemplate{
		subject: document.Subject, text: document.Text, html: document.HTML,
	}, nil
}

func validInvitationBody(value string) bool {
	return len(value) >= len(invitationURLPlaceholder) && len(value) <= 32*1024 &&
		strings.Count(value, invitationURLPlaceholder) == 1 &&
		!strings.ContainsRune(value, 0)
}

func (template IAMInvitationTemplate) Values() (
	subject string,
	textBody string,
	htmlBody string,
	err error,
) {
	if template.subject == "" || !validInvitationBody(template.text) ||
		!validInvitationBody(template.html) {
		return "", "", "", errInvalidIAMWorkerConfig
	}
	return template.subject, template.text, template.html, nil
}
