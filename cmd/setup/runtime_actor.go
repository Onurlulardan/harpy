package main

import (
	"crypto/ed25519"
	"encoding/json"
)

type actorBridgePublicDocument struct {
	Version int                    `json:"version"`
	Keys    []actorBridgePublicKey `json:"keys"`
}

type actorBridgePublicKey struct {
	KeyID     string `json:"key_id"`
	PublicKey string `json:"public_key"`
}

type actorBridgeTargetsDocument struct {
	Version int                      `json:"version"`
	Targets []actorBridgeTargetEntry `json:"targets"`
}

type actorBridgeTargetEntry struct {
	Domain     string `json:"domain"`
	BaseURL    string `json:"base_url"`
	ServerName string `json:"server_name"`
}

type invitationTemplateDocument struct {
	Version int    `json:"version"`
	Subject string `json:"subject"`
	Text    string `json:"text"`
	HTML    string `json:"html"`
}

func writeActorBridge(dir string) ([]string, error) {
	seed, err := markedMaterial(21)
	if err != nil {
		return nil, err
	}
	private := ed25519.NewKeyFromSeed(seed[:])
	const keyID = "edge-local"
	publicDocument, err := json.Marshal(actorBridgePublicDocument{
		Version: 1,
		Keys: []actorBridgePublicKey{{
			KeyID: keyID, PublicKey: rawKey(private[ed25519.SeedSize:]),
		}},
	})
	if err != nil {
		return nil, err
	}
	targets, err := json.Marshal(actorBridgeTargetsDocument{
		Version: 1, Targets: localActorBridgeTargets(),
	})
	if err != nil {
		return nil, err
	}
	derived, err := writeRuntimeGroup(dir, map[string][]byte{
		"actor-bridge-seed":             []byte(rawKey(seed[:])),
		"actor-bridge-public-keys.json": publicDocument,
	})
	if err != nil {
		return nil, err
	}
	seedPath := derived["actor-bridge-seed"]
	keysPath := derived["actor-bridge-public-keys.json"]
	targetsPath, err := writeRuntimeFile(dir, "actor-bridge-targets.json", targets)
	if err != nil {
		return nil, err
	}
	certPath, err := writeRuntimeFile(dir, "actor-bridge-client.crt", []byte("local client certificate"))
	if err != nil {
		return nil, err
	}
	keyPath, err := writeRuntimeFile(dir, "actor-bridge-client.key", []byte("local client private key"))
	if err != nil {
		return nil, err
	}
	caPath, err := writeRuntimeFile(dir, "actor-bridge-server-ca.crt", []byte("local server CA"))
	if err != nil {
		return nil, err
	}
	return []string{
		runtimeLine("HARPY_IAM_V1_ACTOR_BRIDGE_SIGNING_KEY_ID", keyID),
		runtimeLine("HARPY_IAM_V1_ACTOR_BRIDGE_SIGNING_SEED_FILE", seedPath),
		runtimeLine("HARPY_IAM_V1_ACTOR_BRIDGE_PUBLIC_KEYS_FILE", keysPath),
		runtimeLine("HARPY_IAM_V1_ACTOR_BRIDGE_TARGETS_FILE", targetsPath),
		runtimeLine("HARPY_IAM_V1_ACTOR_BRIDGE_CLIENT_CERT_FILE", certPath),
		runtimeLine("HARPY_IAM_V1_ACTOR_BRIDGE_CLIENT_KEY_FILE", keyPath),
		runtimeLine("HARPY_IAM_V1_ACTOR_BRIDGE_SERVER_CA_FILE", caPath),
	}, nil
}

func localActorBridgeTargets() []actorBridgeTargetEntry {
	names := []string{
		"analytics", "billing", "catalog", "documents", "notifications", "work",
	}
	targets := make([]actorBridgeTargetEntry, 0, len(names))
	for _, name := range names {
		host := "v1-" + name + ".invalid"
		targets = append(targets, actorBridgeTargetEntry{
			Domain: name, BaseURL: "https://" + host, ServerName: host,
		})
	}
	return targets
}

func writeInvitationMaterial(dir string) ([]string, error) {
	password, err := markedMaterial(22)
	if err != nil {
		return nil, err
	}
	passwordPath, err := writeRuntimeFile(dir, "invitation-smtp-password", []byte(rawKey(password[:])))
	if err != nil {
		return nil, err
	}
	template, err := json.Marshal(invitationTemplateDocument{
		Version: 1, Subject: "Invitation",
		Text: "Open {{INVITATION_URL}}",
		HTML: `<a href="{{INVITATION_URL}}">Open</a>`,
	})
	if err != nil {
		return nil, err
	}
	templatePath, err := writeRuntimeFile(dir, "invitation-template.json", template)
	if err != nil {
		return nil, err
	}
	return []string{
		runtimeLine("HARPY_IAM_INVITATION_SMTP_HOST", "localhost"),
		runtimeLine("HARPY_IAM_INVITATION_SMTP_PORT", "587"),
		runtimeLine("HARPY_IAM_INVITATION_SMTP_USERNAME", "harpy"),
		runtimeLine("HARPY_IAM_INVITATION_SMTP_PASSWORD_FILE", passwordPath),
		runtimeLine("HARPY_IAM_INVITATION_FROM_ADDRESS", "security@localhost"),
		runtimeLine("HARPY_IAM_INVITATION_FROM_NAME", "Harpy Security"),
		runtimeLine("HARPY_IAM_INVITATION_PUBLIC_BASE_URL", "http://127.0.0.1:8080"),
		runtimeLine("HARPY_IAM_INVITATION_TEMPLATE_FILE", templatePath),
		runtimeLine("HARPY_IAM_INVITATION_SMTP_TIMEOUT", "5s"),
	}, nil
}
