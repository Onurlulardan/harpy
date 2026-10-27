package worker

import (
	"fmt"

	iammodule "github.com/Onurlulardan/harpy/core/internal/modules/iam"
	"github.com/Onurlulardan/harpy/core/internal/platform/async/jobexecution"
)

func appendIAMTransferProcessors(
	processors []jobexecution.Processor,
	imports jobexecution.Processor,
	exports []jobexecution.Processor,
) ([]jobexecution.Processor, error) {
	if imports == nil || len(exports) != 2 || exports[0] == nil || exports[1] == nil ||
		iammodule.ValidateSystemImportJobPolicy(imports.Policy()) != nil {
		return nil, fmt.Errorf("IAM transfer processor registry is incomplete")
	}
	foundExport, foundReconcile := false, false
	for _, processor := range exports {
		policy := processor.Policy()
		switch policy.Kind {
		case iammodule.SystemExportJobKind:
			if foundExport || iammodule.ValidateSystemExportJobPolicy(policy) != nil {
				return nil, fmt.Errorf("IAM export processor registry is invalid")
			}
			foundExport = true
		case iammodule.SystemExportReconcileJobKind:
			if foundReconcile ||
				iammodule.ValidateSystemExportReconcileJobPolicy(policy) != nil {
				return nil, fmt.Errorf("IAM export reconciliation registry is invalid")
			}
			foundReconcile = true
		default:
			return nil, fmt.Errorf("IAM export processor registry has an unknown kind")
		}
	}
	if !foundExport || !foundReconcile {
		return nil, fmt.Errorf("IAM export processor registry is incomplete")
	}
	result := append([]jobexecution.Processor(nil), processors...)
	result = append(result, imports)
	return append(result, exports...), nil
}
