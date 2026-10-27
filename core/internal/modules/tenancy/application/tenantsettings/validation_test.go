package tenantsettings

import (
	"bytes"
	"testing"
)

func TestUpdateDigestFencesRevisionAndSMTPSecret(t *testing.T) {
	name := "Acme"
	base := UpdateCommand{Name: &name, ExpectedRevision: 3,
		IdempotencyKey: "tenant-settings-0001", SMTP: &SMTPMutation{
			Host: "smtp.example.test", Port: 587, Username: "mailer",
			Password: []byte("secret-one"), From: "noreply@example.test"}}
	same := base
	same.SMTP = &SMTPMutation{Host: base.SMTP.Host, Port: base.SMTP.Port,
		Username: base.SMTP.Username, Password: append([]byte(nil), base.SMTP.Password...),
		From: base.SMTP.From}
	if updateDigest(base) != updateDigest(same) {
		t.Fatal("same canonical update produced a different digest")
	}
	same.ExpectedRevision++
	if updateDigest(base) == updateDigest(same) {
		t.Fatal("revision was not fenced into the update digest")
	}
	same.ExpectedRevision = base.ExpectedRevision
	same.SMTP.Password = []byte("secret-two")
	if updateDigest(base) == updateDigest(same) {
		t.Fatal("SMTP secret was not fenced into the update digest")
	}
}

func TestUpdateClearsSMTPSecretOnRejectedApplication(t *testing.T) {
	secret := []byte("smtp-secret")
	command := UpdateCommand{ExpectedRevision: 1,
		IdempotencyKey: "tenant-settings-0001",
		SMTP: &SMTPMutation{Host: "smtp.example.test", Port: 587,
			From: "noreply@example.test", Password: secret}}
	if _, err := (*Application)(nil).Update(t.Context(), platformScopeZero, command); err != ErrInvalid {
		t.Fatalf("Update() error = %v", err)
	}
	if !bytes.Equal(secret, make([]byte, len(secret))) {
		t.Fatal("SMTP secret was not cleared")
	}
}
