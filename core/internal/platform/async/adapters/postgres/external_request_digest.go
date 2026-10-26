package postgres

import (
	canonicaljsonv1 "github.com/Onurlulardan/harpy/contracts/canonicaljson/v1"
	effectsv1 "github.com/Onurlulardan/harpy/contracts/effects/v1"
)

func externalRequestDigest(request effectsv1.Request) (string, error) {
	document, err := request.MarshalJSON()
	if err != nil {
		return "", err
	}
	return canonicaljsonv1.DigestHex(document, effectsv1.MaximumDocumentBytes)
}
