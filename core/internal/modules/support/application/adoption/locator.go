package adoption

import (
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/google/uuid"
)

type SourceKind string

const (
	ContactGlobal SourceKind = "contact_global"
	BugTenant     SourceKind = "bug_tenant"
)

var schemaNamePattern = regexp.MustCompile(`^[a-z][a-z0-9_]{0,62}$`)

type VerifiedLocator struct {
	kind              SourceKind
	tenantID          uuid.UUID
	schema            string
	relation          string
	relationOID       uint32
	tenantSlug        string
	tenantName        string
	attestationSHA256 [32]byte
}

func NewVerifiedLocator(kind SourceKind, tenantID uuid.UUID, schema, relation string,
	relationOID uint32, tenantSlug, tenantName string, attestationSHA256 [32]byte,
) (VerifiedLocator, error) {
	if relationOID == 0 || zeroSHA(attestationSHA256) || !schemaNamePattern.MatchString(schema) {
		return VerifiedLocator{}, ErrInvalid
	}
	locator := VerifiedLocator{kind: kind, tenantID: tenantID, schema: schema,
		relation: relation, relationOID: relationOID, tenantSlug: tenantSlug,
		tenantName: tenantName, attestationSHA256: attestationSHA256}
	switch kind {
	case ContactGlobal:
		if tenantID != uuid.Nil || schema != "public" || relation != "contact_messages" ||
			tenantSlug != "" || tenantName != "" {
			return VerifiedLocator{}, ErrInvalid
		}
	case BugTenant:
		if tenantID == uuid.Nil || relation != "bug_reports" || reservedSchema(schema) ||
			!boundedText(tenantSlug, 1, 100) || !boundedText(tenantName, 1, 200) {
			return VerifiedLocator{}, ErrInvalid
		}
	default:
		return VerifiedLocator{}, ErrInvalid
	}
	return locator, nil
}

func (locator VerifiedLocator) Kind() SourceKind            { return locator.kind }
func (locator VerifiedLocator) TenantID() uuid.UUID         { return locator.tenantID }
func (locator VerifiedLocator) Schema() string              { return locator.schema }
func (locator VerifiedLocator) Relation() string            { return locator.relation }
func (locator VerifiedLocator) RelationOID() uint32         { return locator.relationOID }
func (locator VerifiedLocator) TenantSlug() string          { return locator.tenantSlug }
func (locator VerifiedLocator) TenantName() string          { return locator.tenantName }
func (locator VerifiedLocator) AttestationSHA256() [32]byte { return locator.attestationSHA256 }

func reservedSchema(value string) bool {
	return value == "public" || value == "platform" || value == "information_schema" ||
		value == "pg_catalog" || strings.HasPrefix(value, "pg_")
}

func boundedText(value string, minimum, maximum int) bool {
	return utf8.ValidString(value) && !strings.ContainsRune(value, 0) &&
		utf8.RuneCountInString(value) >= minimum && len([]byte(value)) <= maximum
}

func zeroSHA(value [32]byte) bool { return value == [32]byte{} }
