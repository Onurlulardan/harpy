package local

import (
	"os"
	"strings"
	"testing"
)

func TestRabbitMQConfigFileUsesExtensionlessBasePath(t *testing.T) {
	content, err := os.ReadFile("compose.infra.yml")
	if err != nil {
		t.Fatalf("read compose.infra.yml: %v", err)
	}

	compose := string(content)
	if !strings.Contains(compose, "RABBITMQ_CONFIG_FILE: /run/secrets/rabbitmq\n") {
		t.Fatal("RabbitMQ config must use the extensionless base path")
	}
	if strings.Contains(compose, "RABBITMQ_CONFIG_FILE: /run/secrets/rabbitmq.conf") {
		t.Fatal("RabbitMQ appends .conf; an explicit extension resolves to .conf.conf")
	}
}

func TestClamAVUpdatesSignaturesBeforeServing(t *testing.T) {
	content, err := os.ReadFile("compose.infra.yml")
	if err != nil {
		t.Fatalf("read compose.infra.yml: %v", err)
	}

	compose := string(content)
	required := []string{
		"  clamav-update:\n",
		"command: [\"/bin/sh\", \"/scripts/update-clamav.sh\"]",
		"CLAMAV_NO_FRESHCLAMD: \"true\"",
		"test: [\"CMD\", \"clamdscan\", \"--ping\", \"3:1\"]",
	}
	for _, contract := range required {
		if !strings.Contains(compose, contract) {
			t.Fatalf("ClamAV update contract is missing %q", contract)
		}
	}
	clamav := composeServiceBlock(t, compose, "clamav")
	if strings.Contains(clamav, "clamav-update:") {
		t.Fatal("ClamAV daemon must not traverse the lifecycle-owned updater")
	}
}

func composeServiceBlock(t *testing.T, document, name string) string {
	t.Helper()
	marker := "\n  " + name + ":\n"
	start := strings.Index(document, marker)
	if start < 0 {
		t.Fatalf("compose service %q is missing", name)
	}
	lines := strings.Split(document[start+1:], "\n")
	end := len(lines)
	for index := 1; index < len(lines); index++ {
		line := lines[index]
		trimmed := strings.TrimLeft(line, " ")
		indent := len(line) - len(trimmed)
		if trimmed != "" && (indent < 2 ||
			(indent == 2 && strings.HasSuffix(trimmed, ":"))) {
			end = index
			break
		}
	}
	return strings.Join(lines[:end], "\n")
}
