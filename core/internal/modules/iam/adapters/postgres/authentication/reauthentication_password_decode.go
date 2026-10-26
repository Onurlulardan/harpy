package postgres

func (row *reauthenticationCredentialRow) decodePassword() error {
	if row.passwordVersion == nil || row.passwordRevision == nil ||
		row.passwordAlgorithm == nil || row.passwordCreatedAt == nil ||
		row.authenticatorID != nil {
		return errAuthenticationStorage
	}
	password, err := decodePasswordCredential(
		*row.passwordAlgorithm, row.passwordSalt, row.passwordTag,
		row.passwordMemory, row.passwordIterations, row.passwordParallelism,
		row.passwordLegacy, row.passwordCreatedAt.UTC(),
	)
	if err != nil {
		return err
	}
	row.credential.PasswordCredentialVersion = *row.passwordVersion
	row.credential.PasswordCredentialRevision = *row.passwordRevision
	row.credential.Password = password
	return nil
}
