package maintenance

import (
	"bytes"
	"crypto/hmac"
	"hash"
	"sort"
)

func mergeDirectoryChecksumSubject(
	target *directoryChecksumSubject,
	row directoryChecksumSubject,
) error {
	if target.id != row.id || !sameDirectorySource(target.source, row.source) {
		row.wipe()
		return ErrRuntimeInvalid
	}
	if row.source != nil {
		row.source.wipe()
	}
	target.actuals = append(target.actuals, row.actuals...)
	row.actuals = nil
	return nil
}

func sameDirectorySource(left, right *directorySourceRow) bool {
	if left == nil || right == nil {
		return left == nil && right == nil
	}
	return left.id == right.id && left.userID == right.userID &&
		left.kind == right.kind && left.state == right.state &&
		left.sourceRevision == right.sourceRevision &&
		bytes.Equal(left.normalizedSHA, right.normalizedSHA) &&
		bytes.Equal(left.normalized, right.normalized)
}

func evaluateDirectoryChecksumSubject(
	subject directoryChecksumSubject,
	keys []DirectoryKey,
	keyCounts map[string]int64,
	digest hash.Hash,
) bool {
	writeDirectoryDigest(digest, []byte(subject.id.String()))
	for _, actual := range subject.actuals {
		keyCounts[actual.keyID]++
		writeDirectoryDigest(digest, []byte(actual.keyID), actual.hmac,
			[]byte(actual.userID.String()), []byte(actual.kind),
			[]byte(actual.state), actual.sourceSHA)
	}
	if subject.source == nil {
		return true
	}
	source := subject.source
	writeDirectoryDigest(digest, []byte(source.userID.String()),
		[]byte(source.kind), []byte(source.state), source.normalizedSHA)
	if source.state != "active" {
		for _, actual := range subject.actuals {
			if !actualMatchesSource(actual, source, "tombstoned") {
				return true
			}
		}
		return false
	}
	if len(subject.actuals) != len(keys) {
		return true
	}
	actuals := make(map[string]directoryActualRow, len(subject.actuals))
	for _, actual := range subject.actuals {
		if _, duplicate := actuals[actual.keyID]; duplicate {
			return true
		}
		actuals[actual.keyID] = actual
	}
	for _, key := range keys {
		actual, found := actuals[key.ID]
		expected := directoryLookupHMAC(key, source.normalized)
		matches := found && actualMatchesSource(actual, source, "active") &&
			hmac.Equal(actual.hmac, expected[:])
		writeDirectoryDigest(digest, []byte(key.ID), expected[:])
		clear(expected[:])
		if !matches {
			return true
		}
	}
	return false
}

func actualMatchesSource(
	actual directoryActualRow,
	source *directorySourceRow,
	state string,
) bool {
	return actual.userID == source.userID && actual.kind == source.kind &&
		actual.sourceRevision == source.sourceRevision && actual.state == state &&
		hmac.Equal(actual.sourceSHA, source.normalizedSHA)
}

func directoryKeyEvidence(
	keys []DirectoryKey,
	counts map[string]int64,
) ([]string, []int64) {
	for _, key := range keys {
		if _, found := counts[key.ID]; !found {
			counts[key.ID] = 0
		}
	}
	ids := make([]string, 0, len(counts))
	for id := range counts {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	values := make([]int64, len(ids))
	for index, id := range ids {
		values[index] = counts[id]
	}
	return ids, values
}
