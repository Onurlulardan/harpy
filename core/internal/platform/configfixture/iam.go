package configfixture

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

type EnvironmentSetter interface {
	Helper()
	Setenv(string, string)
}

func SetIAMAPI(setter IAMAPISetter) {
	setter.Helper()
	blocklists, err := passwordBlocklistFixture(setter.TempDir())
	if err != nil {
		setter.Fatalf("create IAM password blocklist fixture: %v", err)
	}
	bridge, err := actorBridgeFixture(setter.TempDir(), time.Now().UTC())
	if err != nil {
		setter.Fatalf("create IAM actor bridge fixture: %v", err)
	}
	for _, entry := range append(iamAPIEntries(time.Now().UTC(), blocklists), bridge...) {
		setter.Setenv(entry.name, entry.value)
	}
}

type IAMWorkerSetter interface {
	EnvironmentSetter
	ArtifactFixture
}

func SetIAMWorker(setter IAMWorkerSetter) {
	setter.Helper()
	entries, err := iamWorkerEntries(time.Now().UTC(), setter.TempDir())
	if err != nil {
		setter.Fatalf("create IAM worker fixture: %v", err)
	}
	for _, entry := range entries {
		setter.Setenv(entry.name, entry.value)
	}
}

func IAMAPIEnvironment(fixture ArtifactFixture, now time.Time) []string {
	fixture.Helper()
	blocklists, err := passwordBlocklistFixture(fixture.TempDir())
	if err != nil {
		fixture.Fatalf("create IAM password blocklist fixture: %v", err)
	}
	bridge, err := actorBridgeFixture(fixture.TempDir(), now)
	if err != nil {
		fixture.Fatalf("create IAM actor bridge fixture: %v", err)
	}
	return environmentStrings(append(iamAPIEntries(now, blocklists), bridge...))
}

func IAMWorkerEnvironment(fixture ArtifactFixture, now time.Time) []string {
	fixture.Helper()
	entries, err := iamWorkerEntries(now, fixture.TempDir())
	if err != nil {
		fixture.Fatalf("create IAM worker fixture: %v", err)
	}
	return environmentStrings(entries)
}

type entry struct {
	name  string
	value string
}

const (
	fixtureTOTPKMSKeyID = "gcp-kms://projects/harpy-fixture/locations/europe-west1/" +
		"keyRings/iam/cryptoKeys/totp/cryptoKeyVersions/7"
	fixtureDeliveryKMSKeyID = "gcp-kms://projects/harpy-fixture/locations/europe-west1/" +
		"keyRings/iam/cryptoKeys/delivery/cryptoKeyVersions/4"
)

func iamAPIEntries(now time.Time, blocklists string) []entry {
	entries := []entry{
		{"DEPLOYMENT_MODE", "saas"},
		{"HARPY_IAM_ACCESS_TOKEN_ISSUER", "https://api.fixture.invalid"},
		{"HARPY_IAM_ACCESS_TOKEN_AUDIENCE", "https://api.fixture.invalid/api"},
		{"HARPY_IAM_ALLOWED_ORIGINS", `["https://app.fixture.invalid"]`},
		{
			name:  "HARPY_IAM_ACCESS_TOKEN_KEYRING",
			value: accessKeyring(now, 10),
		},
	}
	purposes := []struct {
		name    string
		purpose string
		marker  byte
	}{
		{"HARPY_IAM_REFRESH_TOKEN_KEYRING", "refresh_token_v1", 20},
		{"HARPY_IAM_CSRF_KEYRING", "csrf_v1", 21},
		{"HARPY_IAM_LOGIN_DIRECTORY_KEYRING", "login_directory_v1", 22},
		{"HARPY_IAM_AUTH_THROTTLE_KEYRING", "auth_throttle_v1", 23},
		{"HARPY_IAM_SESSION_NETWORK_KEYRING", "session_network_v1", 24},
		{"HARPY_IAM_ONE_TIME_TOKEN_KEYRING", "one_time_token_v1", 25},
		{"HARPY_IAM_RECOVERY_CODE_KEYRING", "recovery_code_v1", 26},
	}
	for _, purpose := range purposes {
		entries = append(entries, entry{
			name: purpose.name, value: hmacKeyring(now, purpose.purpose, purpose.marker),
		})
	}
	return append(entries,
		entry{"HARPY_IAM_PASSWORD_BLOCKLISTS", blocklists},
		entry{"HARPY_IAM_TOTP_KMS_KEY_ID", fixtureTOTPKMSKeyID},
		entry{"HARPY_IAM_DELIVERY_KMS_KEY_ID", fixtureDeliveryKMSKeyID},
	)
}

func iamWorkerEntries(now time.Time, directory string) ([]entry, error) {
	passwordPath := filepath.Join(directory, "smtp-password.txt")
	templatePath := filepath.Join(directory, "invitation-template.json")
	if err := os.WriteFile(passwordPath, []byte("fixture-smtp-password\n"), 0o600); err != nil {
		return nil, err
	}
	template := `{"version":1,"subject":"You're invited to Harpy",` +
		`"text":"Join your Harpy workspace: {{INVITATION_URL}}",` +
		`"html":"<a href=\"{{INVITATION_URL}}\" style=\"background:#18181b;color:#fff;` +
		`padding:12px 28px;text-decoration:none;border-radius:6px\">Accept invitation</a>"}`
	if err := os.WriteFile(templatePath, []byte(template), 0o600); err != nil {
		return nil, err
	}
	return []entry{{
		name: "HARPY_IAM_LOGIN_DIRECTORY_KEYRING",
		value: hmacKeyring(
			now, "login_directory_v1", 22,
		),
	},
		{"HARPY_IAM_ONE_TIME_TOKEN_KEYRING",
			hmacKeyring(now, "one_time_token_v1", 25)},
		{"HARPY_IAM_DELIVERY_KMS_KEY_ID", fixtureDeliveryKMSKeyID},
		{"HARPY_IAM_INVITATION_SMTP_HOST", "smtp.fixture.invalid"},
		{"HARPY_IAM_INVITATION_SMTP_PORT", "587"},
		{"HARPY_IAM_INVITATION_SMTP_USERNAME", "fixture-user"},
		{"HARPY_IAM_INVITATION_SMTP_PASSWORD_FILE", passwordPath},
		{"HARPY_IAM_INVITATION_FROM_ADDRESS", "security@fixture.invalid"},
		{"HARPY_IAM_INVITATION_FROM_NAME", "Harpy Security"},
		{"HARPY_IAM_INVITATION_PUBLIC_BASE_URL", "https://app.fixture.invalid"},
		{"HARPY_IAM_INVITATION_TEMPLATE_FILE", templatePath},
		{"HARPY_IAM_INVITATION_SMTP_TIMEOUT", "5s"},
	}, nil
}

func accessKeyring(now time.Time, marker byte) string {
	activatedAt := canonicalNow(now).Add(-24 * time.Hour)
	return marshal(map[string]any{
		"version": 1,
		"active": map[string]any{
			"kid": keyID(marker), "seed": material(marker),
			"activated_at": timestamp(activatedAt),
			"sign_until":   timestamp(activatedAt.Add(30 * 24 * time.Hour)),
		},
		"retired_public": []any{},
	})
}

func hmacKeyring(now time.Time, purpose string, marker byte) string {
	activatedAt := canonicalNow(now).Add(-24 * time.Hour)
	return marshal(map[string]any{
		"version": 1, "purpose": purpose,
		"active": map[string]any{
			"kid": keyID(marker), "key": material(marker),
			"activated_at": timestamp(activatedAt),
			"issue_until":  timestamp(activatedAt.Add(90 * 24 * time.Hour)),
		},
		"retained": []any{},
	})
}

func environmentStrings(entries []entry) []string {
	values := make([]string, 0, len(entries))
	for _, entry := range entries {
		values = append(values, entry.name+"="+entry.value)
	}
	return values
}

func canonicalNow(value time.Time) time.Time { return value.UTC().Truncate(time.Second) }
func timestamp(value time.Time) string       { return value.Format("2006-01-02T15:04:05Z") }
func keyID(marker byte) string {
	return base64.RawURLEncoding.EncodeToString(bytes.Repeat([]byte{marker}, 12))
}
func material(marker byte) string {
	return base64.RawURLEncoding.EncodeToString(bytes.Repeat([]byte{marker}, 32))
}
func marshal(value any) string {
	document, err := json.Marshal(value)
	if err != nil {
		panic(err)
	}
	return string(document)
}
