package actorassertionv1

import (
	"crypto/sha256"
	"encoding/hex"
	"strconv"
)

const (
	MaximumHTTPBindingPathBytes  = 4096
	MaximumHTTPBindingQueryBytes = 16384
	MaximumHTTPBindingBodyBytes  = int64(1 << 30)
)

// HTTPRequestTargetBindingSHA256 is the shared V2 issuer/V1 consumer profile
// for binding an actor assertion to one exact replayable HTTP request.
func HTTPRequestTargetBindingSHA256(
	method string,
	escapedPath string,
	rawQuery string,
	bodySHA256 [sha256.Size]byte,
	contentLength int64,
) ([sha256.Size]byte, error) {
	if !validHTTPBindingMethod(method) || escapedPath == "" || escapedPath[0] != '/' ||
		len(escapedPath) > MaximumHTTPBindingPathBytes ||
		len(rawQuery) > MaximumHTTPBindingQueryBytes ||
		contentLength < 0 || contentLength > MaximumHTTPBindingBodyBytes {
		return [sha256.Size]byte{}, ErrContract
	}
	target := sha256.Sum256([]byte(escapedPath + "?" + rawQuery))
	return TargetBindingSHA256(
		method,
		hex.EncodeToString(target[:]),
		hex.EncodeToString(bodySHA256[:]),
		strconv.FormatInt(contentLength, 10),
	)
}

func validHTTPBindingMethod(method string) bool {
	switch method {
	case "GET", "HEAD", "POST", "PUT", "PATCH", "DELETE":
		return true
	default:
		return false
	}
}
