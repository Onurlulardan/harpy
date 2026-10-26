package checkpoint

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strconv"
)

const (
	leafDomain = "harpy.protected-authority.checkpoint-leaf.v1\x00"
	nodeDomain = "harpy.protected-authority.checkpoint-node.v1\x00"
)

func hashLeaf(leaf ScopeLeaf) ([32]byte, error) {
	if !validLeaf(leaf) {
		return [32]byte{}, fmt.Errorf("protected authority checkpoint leaf is invalid")
	}
	canonical := canonicalLeaf(leaf)
	return sha256.Sum256(append([]byte(leafDomain), canonical...)), nil
}

func canonicalLeaf(leaf ScopeLeaf) []byte {
	result := []byte(`{"ack_sha256":"` + leaf.AckSHA256 + `","generation":`)
	result = strconv.AppendInt(result, leaf.Generation, 10)
	result = append(result, []byte(`,"registry_binding_id":"`+leaf.RegistryBindingID+
		`","scope_id":"`+leaf.ScopeID+`","scope_kind":"`+leaf.ScopeKind+
		`","state_sha256":"`+leaf.StateSHA256+`"}`)...)
	return result
}

func hashNode(left, right [32]byte) [32]byte {
	message := make([]byte, 0, len(nodeDomain)+64)
	message = append(message, nodeDomain...)
	message = append(message, left[:]...)
	message = append(message, right[:]...)
	return sha256.Sum256(message)
}

func digestString(value [32]byte) string { return hex.EncodeToString(value[:]) }

func parseDigest(value string) ([32]byte, error) {
	var result [32]byte
	decoded, err := hex.DecodeString(value)
	if err != nil || len(decoded) != len(result) {
		return result, fmt.Errorf("checkpoint digest is invalid")
	}
	copy(result[:], decoded)
	if digestString(result) != value {
		return [32]byte{}, fmt.Errorf("checkpoint digest is not canonical")
	}
	return result, nil
}
