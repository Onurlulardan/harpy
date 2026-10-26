package v1query

import (
	"net/url"
	"strings"
	"unicode/utf8"

	restqueryv1 "github.com/Onurlulardan/harpy/contracts/restquery/v1"
)

const maximumV1QueryBytes = 16 * 1024

type wireLeaf struct {
	key   string
	value string
	path  string
}

func decodeWire(rawQuery string) ([]wireLeaf, error) {
	if len(rawQuery) > maximumV1QueryBytes {
		return nil, singleQueryError(
			restqueryv1.ErrorQueryTransportInvalid, "",
			restqueryv1.IssueQuerySizeExceeded,
		)
	}
	if rawQuery == "" {
		return []wireLeaf{}, nil
	}
	leaves := make([]wireLeaf, 0, strings.Count(rawQuery, "&")+1)
	for _, encoded := range strings.Split(rawQuery, "&") {
		if encoded == "" {
			continue
		}
		keyRaw, valueRaw, found := strings.Cut(encoded, "=")
		if !found {
			valueRaw = ""
		}
		key, err := url.QueryUnescape(keyRaw)
		if err != nil {
			return nil, v1PercentError()
		}
		value, err := url.QueryUnescape(valueRaw)
		if err != nil {
			return nil, v1PercentError()
		}
		if !utf8.ValidString(key) || !utf8.ValidString(value) {
			return nil, singleQueryError(
				restqueryv1.ErrorQueryTransportInvalid, "",
				restqueryv1.IssueUTF8Invalid,
			)
		}
		if strings.ContainsRune(key, '\x00') || strings.ContainsRune(value, '\x00') {
			return nil, singleQueryError(
				restqueryv1.ErrorQueryTransportInvalid, "",
				restqueryv1.IssueNULForbidden,
			)
		}
		leaves = append(leaves, wireLeaf{
			key: key, value: value, path: queryKeyPointer(key),
		})
	}
	return leaves, nil
}

func v1PercentError() error {
	return singleQueryError(
		restqueryv1.ErrorQueryTransportInvalid, "",
		restqueryv1.IssuePercentEncodingInvalid,
	)
}
