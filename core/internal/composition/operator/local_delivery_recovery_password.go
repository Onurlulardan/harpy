package operator

import (
	"os"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam"
	"github.com/Onurlulardan/harpy/core/internal/platform/installationsecret"
)

func prepareRecoveryPassword(path string) (iam.PreparedInstallationCredential, error) {
	if _, err := os.Lstat(path); os.IsNotExist(err) {
		return createPreparedPassword(path)
	} else if err != nil {
		return iam.PreparedInstallationCredential{}, err
	}
	var prepared iam.PreparedInstallationCredential
	err := installationsecret.Read(path, func(wire []byte) error {
		var prepareErr error
		prepared, prepareErr = iam.PrepareInstallationCredential(wire)
		return prepareErr
	})
	return prepared, err
}

func removePasswordIfPresent(path string) error {
	if _, err := os.Lstat(path); os.IsNotExist(err) {
		return nil
	} else if err != nil {
		return err
	}
	return installationsecret.Remove(path)
}
