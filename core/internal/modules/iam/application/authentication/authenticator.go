package authentication

import (
	"reflect"
)

type Authenticator struct {
	port        Port
	password    PasswordCryptography
	refresh     RefreshGenerator
	keys        RefreshKeyring
	access      AccessIssuer
	security    SecurityCryptography
	identifiers UUIDSource
	clock       Clock
	valid       bool
}

func NewAuthenticator(
	port Port,
	password PasswordCryptography,
	refresh RefreshGenerator,
	keys RefreshKeyring,
	access AccessIssuer,
	security SecurityCryptography,
	identifiers UUIDSource,
	clock Clock,
) (Authenticator, error) {
	if nilInterface(port) || nilInterface(password) || nilInterface(refresh) ||
		nilInterface(keys) || nilInterface(access) || identifiers == nil || clock == nil {
		return Authenticator{}, ErrAuthenticationContract
	}
	if security.validate() != nil {
		return Authenticator{}, ErrAuthenticationContract
	}
	return Authenticator{
		port: port, password: password, refresh: refresh, keys: keys,
		access: access, security: security, identifiers: identifiers,
		clock: clock, valid: true,
	}, nil
}

func nilInterface(value any) bool {
	if value == nil {
		return true
	}
	seen := reflect.ValueOf(value)
	switch seen.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface,
		reflect.Map, reflect.Pointer, reflect.Slice:
		return seen.IsNil()
	default:
		return false
	}
}
