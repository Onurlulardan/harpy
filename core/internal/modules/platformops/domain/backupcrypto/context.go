package backupcrypto

func validContext(value Context) bool {
	return validUUID(value.EnvironmentID) && validUUID(value.RecoverySetID) &&
		componentPattern.MatchString(value.ComponentID) &&
		validKeyID(value.DRKeyID) && validKeyID(value.RecoveryKeyID)
}

func wrapInfo(value Context) []byte {
	return []byte("harpy.fcbackup1.wrap.v1\x00" + value.EnvironmentID + "\x00" +
		value.RecoverySetID + "\x00" + value.ComponentID)
}

func (value archiveHeader) context() Context {
	return Context{EnvironmentID: value.EnvironmentID,
		RecoverySetID: value.RecoverySetID, ComponentID: value.ComponentID,
		DRKeyID: value.DRWrap.KeyID, RecoveryKeyID: value.RecoveryWrap.KeyID}
}
