package adoption

import (
	"crypto/sha256"
	"testing"

	"github.com/google/uuid"
)

func TestVerifiedLocatorPinsSourceKindTenantAndCatalogIdentity(t *testing.T) {
	digest := sha256.Sum256([]byte("locator"))
	contact, err := NewVerifiedLocator(ContactGlobal, uuid.Nil, "public",
		"contact_messages", 42, "", "", digest)
	if err != nil || contact.Kind() != ContactGlobal || contact.RelationOID() != 42 {
		t.Fatalf("contact locator=%+v err=%v", contact, err)
	}
	tenantID := uuid.New()
	bug, err := NewVerifiedLocator(BugTenant, tenantID, "tenant_acme", "bug_reports",
		84, "acme", "Acme", digest)
	if err != nil || bug.TenantID() != tenantID || bug.Schema() != "tenant_acme" {
		t.Fatalf("bug locator=%+v err=%v", bug, err)
	}
	for name, input := range map[string]struct {
		kind     SourceKind
		tenantID uuid.UUID
		schema   string
		relation string
		digest   [32]byte
	}{
		"contact tenant": {ContactGlobal, tenantID, "public", "contact_messages", digest},
		"reserved bug":   {BugTenant, tenantID, "platform", "bug_reports", digest},
		"raw relation":   {BugTenant, tenantID, "tenant_acme", "arbitrary_table", digest},
		"missing proof":  {BugTenant, tenantID, "tenant_acme", "bug_reports", [32]byte{}},
	} {
		t.Run(name, func(t *testing.T) {
			_, createErr := NewVerifiedLocator(input.kind, input.tenantID, input.schema,
				input.relation, 84, "acme", "Acme", input.digest)
			if createErr == nil {
				t.Fatal("unsafe source locator was accepted")
			}
		})
	}
}
