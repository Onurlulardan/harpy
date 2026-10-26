package catalogpackagev1

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"

	canonicaljsonv1 "github.com/Onurlulardan/harpy/contracts/canonicaljson/v1"
)

func DecodeCanonical(document []byte) (ReleaseSet, error) {
	canonical, err := canonicaljsonv1.CanonicalizeObject(
		document, MaximumPackageSetBytes,
	)
	if err != nil || !bytes.Equal(canonical, document) {
		return ReleaseSet{}, fmt.Errorf("catalog package set is not canonical")
	}
	var set ReleaseSet
	decoder := json.NewDecoder(bytes.NewReader(document))
	decoder.DisallowUnknownFields()
	if decoder.Decode(&set) != nil || decoder.Decode(&struct{}{}) != io.EOF {
		return ReleaseSet{}, fmt.Errorf("catalog package set contract is invalid")
	}
	if err := set.Validate(); err != nil {
		return ReleaseSet{}, err
	}
	return set.Clone(), nil
}
