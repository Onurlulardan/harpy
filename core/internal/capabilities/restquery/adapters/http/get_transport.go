package httpquery

import (
	"net/url"
	"strings"
	"unicode/utf8"

	restqueryv1 "github.com/Onurlulardan/harpy/contracts/restquery/v1"
	"golang.org/x/text/unicode/norm"
)

type getLeaf struct {
	segments      []string
	pointer       string
	value         string
	structuralNFC bool
}

func decodeGETLeaves(rawQuery string) ([]getLeaf, error) {
	if rawQuery == "" {
		return []getLeaf{}, nil
	}
	if strings.ContainsRune(rawQuery, '+') {
		return nil, getTransportError(restqueryv1.IssueRawPlusForbidden)
	}
	if strings.ContainsRune(rawQuery, ' ') {
		return nil, getTransportError(restqueryv1.IssuePercentEncodingInvalid)
	}
	leaves := make([]getLeaf, 0, strings.Count(rawQuery, "&")+1)
	seen := make(map[string]struct{}, cap(leaves))
	duplicates := make(map[string]struct{})
	for _, encoded := range strings.Split(rawQuery, "&") {
		leaf, canonicalName, err := decodeGETLeaf(encoded)
		if err != nil {
			return nil, err
		}
		if _, found := seen[canonicalName]; found {
			duplicates[leaf.pointer] = struct{}{}
		} else {
			seen[canonicalName] = struct{}{}
		}
		leaves = append(leaves, leaf)
	}
	if len(duplicates) != 0 {
		collector := issueCollector{}
		for pointer := range duplicates {
			collector.append(pointer, restqueryv1.IssueMemberDuplicate)
		}
		return nil, queryError(restqueryv1.ErrorQueryDuplicateMember, collector.issues)
	}
	return leaves, nil
}

func decodeGETLeaf(encoded string) (getLeaf, string, error) {
	if strings.Count(encoded, "=") != 1 {
		return getLeaf{}, "", singleQueryError(
			restqueryv1.ErrorQueryInvalid, "", restqueryv1.IssueValueNotCanonical,
		)
	}
	nameRaw, valueRaw, _ := strings.Cut(encoded, "=")
	name, err := decodeGETComponent(nameRaw)
	if err != nil {
		return getLeaf{}, "", err
	}
	value, err := decodeGETComponent(valueRaw)
	if err != nil {
		return getLeaf{}, "", err
	}
	canonicalName := norm.NFC.String(name)
	segments, validName := splitGETName(name)
	pointer := getNamePointer(canonicalName)
	return getLeaf{
		segments: segments, pointer: pointer, value: value,
		structuralNFC: validName && getSegmentsAreNFC(segments),
	}, canonicalName, nil
}

func decodeGETComponent(encoded string) (string, error) {
	decoded, err := url.PathUnescape(encoded)
	if err != nil {
		return "", getTransportError(restqueryv1.IssuePercentEncodingInvalid)
	}
	if !utf8.ValidString(decoded) {
		return "", getTransportError(restqueryv1.IssueUTF8Invalid)
	}
	if strings.ContainsRune(decoded, '\x00') {
		return "", getTransportError(restqueryv1.IssueNULForbidden)
	}
	return decoded, nil
}
