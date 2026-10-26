package postgres

import "regexp"

var (
	publisherDigestPattern = regexp.MustCompile(`^[0-9a-f]{64}$`)
	publisherCodePattern   = regexp.MustCompile(`^[A-Z][A-Z0-9_]{1,63}$`)
)
