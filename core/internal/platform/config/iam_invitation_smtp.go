package config

import (
	"net/mail"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type IAMInvitationSMTP struct {
	iamKeyRedaction
	host     string
	port     int
	username string
	password Secret
	from     string
	fromName string
	baseURL  string
	timeout  time.Duration
	template IAMInvitationTemplate
	valid    bool
}

func loadIAMInvitationSMTP(
	values environmentValues,
	environment Environment,
) (IAMInvitationSMTP, error) {
	host, err := requiredSMTPText(values, "HARPY_IAM_INVITATION_SMTP_HOST", 253)
	if err != nil || !validSMTPHost(host) {
		return IAMInvitationSMTP{}, configError("HARPY_IAM_INVITATION_SMTP_HOST", "is invalid")
	}
	portValue, err := parseInteger(values, "HARPY_IAM_INVITATION_SMTP_PORT", "587", 1, 65535)
	if err != nil {
		return IAMInvitationSMTP{}, err
	}
	username, err := requiredSMTPText(values, "HARPY_IAM_INVITATION_SMTP_USERNAME", 320)
	if err != nil {
		return IAMInvitationSMTP{}, err
	}
	passwordPath, exists := values.lookup("HARPY_IAM_INVITATION_SMTP_PASSWORD_FILE")
	if !exists {
		return IAMInvitationSMTP{}, configError("HARPY_IAM_INVITATION_SMTP_PASSWORD_FILE", "is required")
	}
	passwordValue, err := readSecretFile(
		"HARPY_IAM_INVITATION_SMTP_PASSWORD", passwordPath, 4096,
	)
	if err != nil || validateSecretValue("HARPY_IAM_INVITATION_SMTP_PASSWORD", passwordValue, 4096) != nil {
		return IAMInvitationSMTP{}, configError("HARPY_IAM_INVITATION_SMTP_PASSWORD_FILE", "is invalid")
	}
	from, err := requiredSMTPText(values, "HARPY_IAM_INVITATION_FROM_ADDRESS", 320)
	if err != nil || !validMailbox(from) {
		return IAMInvitationSMTP{}, configError("HARPY_IAM_INVITATION_FROM_ADDRESS", "is invalid")
	}
	fromName, err := requiredSMTPText(values, "HARPY_IAM_INVITATION_FROM_NAME", 128)
	if err != nil {
		return IAMInvitationSMTP{}, err
	}
	baseURL, err := loadInvitationBaseURL(values, environment)
	if err != nil {
		return IAMInvitationSMTP{}, err
	}
	timeout, err := parseDuration(
		values, "HARPY_IAM_INVITATION_SMTP_TIMEOUT", "10s",
		time.Second, 30*time.Second,
	)
	if err != nil {
		return IAMInvitationSMTP{}, err
	}
	template, err := loadIAMInvitationTemplate(values)
	if err != nil {
		return IAMInvitationSMTP{}, err
	}
	return IAMInvitationSMTP{
		host: host, port: int(portValue), username: username,
		password: Secret{value: passwordValue}, from: from, fromName: fromName,
		baseURL: baseURL, timeout: timeout, template: template, valid: true,
	}, nil
}

func requiredSMTPText(values environmentValues, key string, maximum int) (string, error) {
	value, exists := values.lookup(key)
	if !exists || value == "" || len(value) > maximum || strings.TrimSpace(value) != value ||
		strings.ContainsAny(value, "\x00\r\n") {
		return "", configError(key, "is required and must be canonical")
	}
	return value, nil
}

func validSMTPHost(value string) bool {
	if strings.Contains(value, "..") || strings.ContainsAny(value, ":/[]@") {
		return false
	}
	for _, label := range strings.Split(value, ".") {
		if label == "" || len(label) > 63 || label[0] == '-' || label[len(label)-1] == '-' {
			return false
		}
		for _, character := range label {
			if character != '-' && (character < '0' || character > '9') &&
				(character < 'A' || character > 'Z') && (character < 'a' || character > 'z') {
				return false
			}
		}
	}
	return true
}

func validMailbox(value string) bool {
	parsed, err := mail.ParseAddress(value)
	return err == nil && parsed.Name == "" && parsed.Address == value
}

func loadInvitationBaseURL(values environmentValues, environment Environment) (string, error) {
	raw, err := requiredSMTPText(values, "HARPY_IAM_INVITATION_PUBLIC_BASE_URL", 2048)
	if err != nil {
		return "", err
	}
	parsed, err := url.Parse(raw)
	validScheme := parsed != nil && (parsed.Scheme == "https" ||
		(parsed.Scheme == "http" && !environment.productionLike()))
	if err != nil || !validScheme || parsed.Host == "" || parsed.User != nil ||
		parsed.RawQuery != "" || parsed.Fragment != "" || parsed.Path != "" ||
		parsed.String() != raw || strings.HasSuffix(raw, "/") {
		return "", configError("HARPY_IAM_INVITATION_PUBLIC_BASE_URL", "is invalid")
	}
	return raw, nil
}

func (config IAMInvitationSMTP) Validate() error {
	if !config.valid || !validSMTPHost(config.host) || config.port < 1 || config.port > 65535 ||
		config.username == "" || config.password.Value() == "" || !validMailbox(config.from) ||
		config.fromName == "" || config.baseURL == "" || config.timeout < time.Second ||
		config.timeout > 30*time.Second {
		return errInvalidIAMWorkerConfig
	}
	_, _, _, err := config.template.Values()
	return err
}

func (config IAMInvitationSMTP) Connection() (string, int, string, Secret, time.Duration, error) {
	if err := config.Validate(); err != nil {
		return "", 0, "", Secret{}, 0, err
	}
	return config.host, config.port, config.username, config.password, config.timeout, nil
}

func (config IAMInvitationSMTP) Message() (string, string, string, IAMInvitationTemplate, error) {
	if err := config.Validate(); err != nil {
		return "", "", "", IAMInvitationTemplate{}, err
	}
	return config.from, config.fromName, config.baseURL, config.template, nil
}

func (config IAMInvitationSMTP) PortString() string { return strconv.Itoa(config.port) }
