package postgres

import (
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/authentication"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/password"
)

func decodePasswordCredential(
	algorithm string,
	salt, tag []byte,
	memory *int32,
	iterations, parallelism *int16,
	legacy *string,
	createdAt time.Time,
) (authentication.PasswordCredential, error) {
	switch algorithm {
	case authentication.AlgorithmArgon2idV1:
		if memory == nil || iterations == nil || parallelism == nil || legacy != nil {
			return authentication.PasswordCredential{}, errAuthenticationStorage
		}
		credential, err := password.Argon2idCredentialFromStorage(
			salt, tag, uint32(*memory), uint32(*iterations), uint8(*parallelism),
		)
		if err != nil {
			return authentication.PasswordCredential{}, errAuthenticationStorage
		}
		return authentication.PasswordCredential{
			Algorithm: algorithm, Argon2id: credential,
		}, nil
	case authentication.AlgorithmBcryptV1Adopted:
		if len(salt) != 0 || len(tag) != 0 || memory != nil || iterations != nil ||
			parallelism != nil || legacy == nil {
			return authentication.PasswordCredential{}, errAuthenticationStorage
		}
		credential, err := password.LegacyBcryptCredentialFromStorage(*legacy, createdAt)
		if err != nil {
			return authentication.PasswordCredential{}, errAuthenticationStorage
		}
		return authentication.PasswordCredential{
			Algorithm: algorithm, Legacy: credential,
		}, nil
	default:
		return authentication.PasswordCredential{}, errAuthenticationStorage
	}
}
