package campaign

import "regexp"

var keyPattern = regexp.MustCompile(`^[a-z][a-z0-9._-]{0,126}[a-z0-9]$`)

type Key string

func NewKey(value string) (Key, error) {
	if !keyPattern.MatchString(value) {
		return "", ErrKey
	}
	return Key(value), nil
}

func (key Key) Valid() bool {
	return keyPattern.MatchString(string(key))
}
