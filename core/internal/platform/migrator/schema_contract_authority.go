package migrator

import (
	runtimev1 "github.com/Onurlulardan/harpy/contracts/runtime/v1"
	"github.com/Onurlulardan/harpy/core/internal/platform/schemacontract"
)

type schemaContractAuthority interface {
	Contract() runtimev1.BinarySchemaContract
}

type releaseSchemaAuthority struct {
	binding schemacontract.ReleaseBinding
}

func (authority releaseSchemaAuthority) Contract() runtimev1.BinarySchemaContract {
	return authority.binding.Contract().Value()
}
