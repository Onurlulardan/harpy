package main

import (
	"encoding/base64"
	"fmt"
)

func buildRabbitMQConfig(password string) (string, error) {
	quotedPassword, err := quoteRabbitMQGeneratedPassword(password)
	if err != nil {
		return "", err
	}
	return "default_user = harpy_worker\n" +
		"default_pass = " + quotedPassword + "\n" +
		"default_vhost = harpy\n", nil
}

func quoteRabbitMQGeneratedPassword(value string) (string, error) {
	decoded, err := base64.RawURLEncoding.DecodeString(value)
	valid := err == nil && len(decoded) == 32 &&
		base64.RawURLEncoding.EncodeToString(decoded) == value
	clear(decoded)
	if !valid {
		return "", fmt.Errorf("rabbitmq password must be canonical base64url")
	}
	return "'" + value + "'", nil
}
