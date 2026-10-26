package effectexecution

import (
	"fmt"
	"regexp"
)

var identityPattern = regexp.MustCompile(`^[a-z][a-z0-9_.-]{2,127}$`)

type Identity struct {
	provider  string
	operation string
}

func NewIdentity(provider string, operation string) (Identity, error) {
	identity := Identity{provider: provider, operation: operation}
	if !identity.valid() {
		return Identity{}, fmt.Errorf("effect provider identity is invalid")
	}
	return identity, nil
}

func (identity Identity) Provider() string  { return identity.provider }
func (identity Identity) Operation() string { return identity.operation }

func (identity Identity) valid() bool {
	return identityPattern.MatchString(identity.provider) &&
		identityPattern.MatchString(identity.operation)
}

func (identity Identity) key() string {
	return identity.provider + "\x00" + identity.operation
}
