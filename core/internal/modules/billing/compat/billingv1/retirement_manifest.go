package billingv1

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"github.com/Onurlulardan/harpy/core/internal/platform/contractretirement"
)

const (
	retirementOwner  = "billing_v1_http"
	retirementDomain = "billing"
	retirementKeyTag = "harpy/billing/v1-http-route-retirement/v1\x00"
)

func RetirementIdentity() (contractretirement.Identity, error) {
	return contractretirement.NewIdentity(retirementOwner, retirementDomain)
}

func RetirementKey(route Route) (contractretirement.Key, error) {
	if route.ID == "" || route.Method == "" || route.LegacyPath == "" {
		return contractretirement.Key{}, fmt.Errorf("Billing V1 route identity is invalid")
	}
	digest := sha256.Sum256([]byte(
		retirementKeyTag + route.ID + "\x00" + route.Method + "\x00" + route.LegacyPath,
	))
	return contractretirement.ParseKey("rq.alias." + hex.EncodeToString(digest[:]))
}

func RetirementManifest() (contractretirement.Batch, error) {
	identity, err := RetirementIdentity()
	if err != nil {
		return contractretirement.Batch{}, err
	}
	routes := Routes()
	keys := make([]contractretirement.Key, len(routes))
	for index, route := range routes {
		keys[index], err = RetirementKey(route)
		if err != nil {
			return contractretirement.Batch{}, err
		}
	}
	return contractretirement.NewBatch(identity, keys)
}
