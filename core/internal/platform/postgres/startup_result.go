package postgres

import runtimev1 "github.com/Onurlulardan/harpy/contracts/runtime/v1"

type StartupResult struct {
	profile              Profile
	applicationName      string
	serverVersion        int
	databaseOID          uint32
	databaseName         string
	systemIdentifier     string
	vectorSchema         string
	databaseOrigin       runtimev1.DatabaseOrigin
	binaryContractDigest string
}

func (result StartupResult) Profile() Profile         { return result.profile }
func (result StartupResult) ApplicationName() string  { return result.applicationName }
func (result StartupResult) ServerVersionNumber() int { return result.serverVersion }
func (result StartupResult) DatabaseOID() uint32      { return result.databaseOID }
func (result StartupResult) DatabaseName() string     { return result.databaseName }
func (result StartupResult) SystemIdentifier() string { return result.systemIdentifier }
func (result StartupResult) VectorSchema() string     { return result.vectorSchema }
func (result StartupResult) DatabaseOrigin() runtimev1.DatabaseOrigin {
	return result.databaseOrigin
}
func (result StartupResult) BinaryContractDigest() string {
	return result.binaryContractDigest
}
