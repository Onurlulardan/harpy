package gcpkms

import kmsv1 "github.com/Onurlulardan/harpy/contracts/kms/v1"

type keyVersion = kmsv1.GoogleCloudKeyVersion

func parseKeyVersion(value string) (keyVersion, error) {
	version, err := kmsv1.ParseGoogleCloudKeyVersion(value)
	if err != nil {
		return keyVersion{}, ErrContract
	}
	return version, nil
}
