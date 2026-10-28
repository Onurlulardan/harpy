package main

import (
	"strings"
	"testing"
)

func TestRabbitMQConfigUsesGeneratedPasswordWithoutInjection(t *testing.T) {
	set, _ := generatedSetForTest(t)
	password := string(set.files["rabbitmq-password.txt"])
	want := "default_user = harpy_worker\n" +
		"default_pass = '" + password + "'\n" +
		"default_vhost = harpy\n"
	if got := string(set.files["rabbitmq.conf"]); got != want {
		t.Fatalf("rabbitmq.conf differs:\n%s", got)
	}
	if strings.ContainsAny(password, "'\r\n") {
		t.Fatal("generated RabbitMQ password contains config injection characters")
	}
}

func TestRabbitMQConfigRejectsNonCanonicalOrInjectablePasswords(t *testing.T) {
	values := []string{"", "not-base64url!", "line\nbreak", "single'quote"}
	for _, value := range values {
		if _, err := buildRabbitMQConfig(value); err == nil {
			t.Fatalf("accepted invalid RabbitMQ password %q", value)
		}
	}
}
