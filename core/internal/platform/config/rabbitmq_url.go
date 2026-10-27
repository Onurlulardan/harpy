package config

import (
	"net/url"
	"strings"
)

func validateRabbitMQURL(key string, secret Secret, environment Environment) error {
	parsed, err := url.Parse(secret.Value())
	if err != nil || parsed == nil || (parsed.Scheme != "amqp" && parsed.Scheme != "amqps") {
		return configError(key, "must be an AMQP URL")
	}
	if environment.productionLike() && parsed.Scheme != "amqps" {
		return configError(key, "must use amqps in this environment")
	}
	password, hasPassword := "", false
	if parsed.User != nil {
		password, hasPassword = parsed.User.Password()
	}
	if parsed.User == nil || parsed.User.Username() == "" || !hasPassword || password == "" ||
		parsed.Hostname() == "" || strings.Contains(parsed.Host, ",") ||
		containsASCIIControl(parsed.Host) || containsASCIIControl(parsed.User.Username()) ||
		containsASCIIControl(password) {
		return configError(key, "must include one host and explicit credentials")
	}
	if parsed.Fragment != "" || parsed.RawQuery != "" ||
		parsed.EscapedPath() == "" || strings.Count(parsed.EscapedPath(), "/") != 1 {
		return configError(key, "must include one vhost and no query or fragment")
	}
	if parsed.Path != "/" && strings.TrimPrefix(parsed.Path, "/") == "" {
		return configError(key, "contains an invalid vhost")
	}
	return nil
}
