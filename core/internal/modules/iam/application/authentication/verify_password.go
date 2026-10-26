package authentication

import "time"

func (authenticator Authenticator) verifyPassword(
	secret PasswordSecret,
	credential PasswordCredential,
	now time.Time,
) (Argon2idCredential, error) {
	var matched bool
	var err error
	switch credential.Algorithm {
	case AlgorithmArgon2idV1:
		matched, err = authenticator.password.VerifyArgon2id(secret, credential.Argon2id)
	case AlgorithmBcryptV1Adopted:
		if !now.Before(credential.Legacy.RetirementDeadline()) {
			return nil, ErrInvalidCredentials
		}
		matched, err = authenticator.password.VerifyLegacy(secret, credential.Legacy, now)
	default:
		return nil, ErrAuthenticationStorage
	}
	if err != nil {
		return nil, ErrAuthenticationCrypto
	}
	if !matched {
		return nil, ErrInvalidCredentials
	}
	if credential.Algorithm != AlgorithmBcryptV1Adopted {
		return nil, nil
	}
	rehash, err := authenticator.password.Hash(secret)
	if err != nil || nilInterface(rehash) || !rehash.Valid() {
		return nil, ErrAuthenticationCrypto
	}
	return rehash, nil
}
