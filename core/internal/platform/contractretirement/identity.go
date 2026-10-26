package contractretirement

import (
	"crypto/sha256"
	"fmt"
	"regexp"
)

const identitySealDomain = "harpy/contract-retirement/identity-seal/v1\x00"

var ownerPattern = regexp.MustCompile(`^[a-z][a-z0-9_]{0,62}$`)

var domains = map[string]struct{}{
	"analytics": {}, "billing": {}, "catalog": {}, "documents": {},
	"forms": {}, "iam": {}, "notifications": {}, "work": {}, "workflow": {},
}

type Identity struct {
	owner  string
	domain string
	seal   [sha256.Size]byte
}

func NewIdentity(owner, domain string) (Identity, error) {
	if !ownerPattern.MatchString(owner) {
		return Identity{}, ErrInvalid
	}
	if _, ok := domains[domain]; !ok {
		return Identity{}, ErrInvalid
	}
	identity := Identity{owner: owner, domain: domain}
	identity.seal = identitySeal(owner, domain)
	return identity, nil
}

func (identity Identity) Validate() error {
	if !ownerPattern.MatchString(identity.owner) ||
		identity.seal != identitySeal(identity.owner, identity.domain) {
		return ErrInvalid
	}
	if _, ok := domains[identity.domain]; !ok {
		return ErrInvalid
	}
	return nil
}

func (identity Identity) DatabaseValues() (string, string, error) {
	if err := identity.Validate(); err != nil {
		return "", "", err
	}
	return identity.owner, identity.domain, nil
}

func identitySeal(owner, domain string) [sha256.Size]byte {
	return sha256.Sum256([]byte(identitySealDomain + owner + "\x00" + domain))
}

func (Identity) String() string                          { return redaction }
func (Identity) GoString() string                        { return redaction }
func (identity Identity) Format(state fmt.State, _ rune) { writeRedaction(state) }
func (Identity) MarshalJSON() ([]byte, error)            { return rejectJSON() }
