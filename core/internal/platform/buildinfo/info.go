package buildinfo

import runtimev1 "github.com/Onurlulardan/harpy/contracts/runtime/v1"

type Info struct {
	version runtimev1.Version
}

func (info Info) Version() runtimev1.Version {
	return info.version
}

func (info Info) Service() runtimev1.Service {
	return info.version.Service
}

func (info Info) Environment() runtimev1.Environment {
	return info.version.Environment
}
