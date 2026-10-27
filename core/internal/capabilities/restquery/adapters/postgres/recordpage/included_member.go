package recordpage

import (
	"bytes"
	"encoding/json"
)

// AppendIncluded appends the reserved member as the final record member.
func AppendIncluded(record []byte, included []byte) ([]byte, error) {
	if !validJSONObject(record) || !validJSONObject(included) ||
		hasIncludedMember(record) {
		return nil, ErrContract
	}
	result := make([]byte, 0, len(record)+len(included)+16)
	result = append(result, record[:len(record)-1]...)
	if len(record) > 2 {
		result = append(result, ',')
	}
	result = append(result, `"_included":`...)
	result = append(result, included...)
	result = append(result, '}')
	if !validJSONObject(result) || !bytes.HasSuffix(
		result, append(append([]byte(`"_included":`), included...), '}'),
	) {
		return nil, ErrContract
	}
	return result, nil
}

func AttachIncluded(page Page, included [][]byte) (Page, error) {
	if page.Validate() != nil || included == nil {
		return Page{}, ErrContract
	}
	records, err := page.Records()
	if err != nil || len(records) != len(included) {
		return Page{}, ErrContract
	}
	for index := range records {
		records[index], err = AppendIncluded(records[index], included[index])
		if err != nil {
			return Page{}, ErrContract
		}
	}
	return newPage(records, page.hasMore, page.nextCursor, page.requestedSize)
}

func hasIncludedMember(document []byte) bool {
	var object map[string]json.RawMessage
	if json.Unmarshal(document, &object) != nil {
		return true
	}
	_, found := object["_included"]
	return found
}
