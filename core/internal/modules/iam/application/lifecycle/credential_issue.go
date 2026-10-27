package lifecycle

import lifecyclecredential "github.com/Onurlulardan/harpy/core/internal/modules/iam/domain/lifecyclecredential"

type IssuedCredential = lifecyclecredential.IssuedCredential

type CredentialGenerator struct {
	generator lifecyclecredential.Generator
}

func NewCredentialGenerator() CredentialGenerator {
	return CredentialGenerator{generator: lifecyclecredential.NewGenerator()}
}

func (generator CredentialGenerator) Generate(
	tenantID [16]byte,
	purpose CredentialPurpose,
	tokenID [16]byte,
	keys OneTimeKeyring,
) (*IssuedCredential, TokenIssue, error) {
	return generator.generator.Generate(
		tenantID, lifecyclecredential.Purpose(purpose), tokenID, keys,
	)
}
