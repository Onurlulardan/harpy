package operator

import (
	"os"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam"
	"github.com/Onurlulardan/harpy/core/internal/platform/installationsecret"
)

func ensurePassword(path string) error {
	if _, err := os.Lstat(path); err == nil {
		return installationsecret.Read(path, func([]byte) error { return nil })
	} else if !os.IsNotExist(err) {
		return err
	}
	password, err := installationsecret.GenerateTemporaryPassword()
	if err != nil {
		return err
	}
	defer password.Destroy()
	return password.Consume(func(wire []byte) error {
		return installationsecret.Publish(path, wire)
	})
}

func createPreparedPassword(path string) (iam.PreparedInstallationCredential, error) {
	password, err := installationsecret.GenerateTemporaryPassword()
	if err != nil {
		return iam.PreparedInstallationCredential{}, err
	}
	defer password.Destroy()
	var prepared iam.PreparedInstallationCredential
	err = password.Consume(func(wire []byte) error {
		var prepareErr error
		prepared, prepareErr = iam.PrepareInstallationCredential(wire)
		if prepareErr != nil {
			return prepareErr
		}
		return installationsecret.Publish(path, wire)
	})
	return prepared, err
}
