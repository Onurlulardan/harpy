package checkpoint

import (
	"regexp"
)

var checkpointDigestPattern = regexp.MustCompile(`^[0-9a-f]{64}$`)
var checkpointUUIDPattern = regexp.MustCompile(
	`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`,
)

func validLeaf(leaf ScopeLeaf) bool {
	return validCheckpointUUID(leaf.RegistryBindingID) &&
		validCheckpointUUID(leaf.ScopeID) && leaf.Generation > 0 &&
		(leaf.ScopeKind == "environment" || leaf.ScopeKind == "tenant") &&
		checkpointDigestPattern.MatchString(leaf.StateSHA256) &&
		checkpointDigestPattern.MatchString(leaf.AckSHA256)
}

func validCheckpointUUID(value string) bool {
	return checkpointUUIDPattern.MatchString(value) &&
		value != "00000000-0000-0000-0000-000000000000"
}

func leafKey(leaf ScopeLeaf) string {
	return leaf.RegistryBindingID + "\x00" + leaf.ScopeKind + "\x00" + leaf.ScopeID
}
