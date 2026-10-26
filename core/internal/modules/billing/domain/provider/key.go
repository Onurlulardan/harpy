package provider

import "regexp"

var keyPattern = regexp.MustCompile(`^[a-z](?:[a-z0-9_]{0,30}[a-z0-9])?$`)

type Key string

const Manual Key = "manual"

func NewKey(value string) (Key, error) {
	key := Key(value)
	if !key.Valid() {
		return "", ErrKey
	}
	return key, nil
}

func (key Key) Valid() bool {
	return keyPattern.MatchString(string(key))
}
