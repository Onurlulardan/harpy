package worker

import (
	"context"
	"testing"

	iammodule "github.com/Onurlulardan/harpy/core/internal/modules/iam"
	asyncpostgres "github.com/Onurlulardan/harpy/core/internal/platform/async/adapters/postgres"
	"github.com/Onurlulardan/harpy/core/internal/platform/async/jobexecution"
)

type iamTransferProcessorStub struct{ policy jobexecution.Policy }

func (processor iamTransferProcessorStub) Policy() jobexecution.Policy {
	return processor.policy
}

func (iamTransferProcessorStub) Process(
	context.Context,
	asyncpostgres.JobClaim,
) error {
	return nil
}

func TestProductionIAMTransferCompositionCoversReleaseRegistry(t *testing.T) {
	imports := iamTransferProcessorStub{policy: iammodule.SystemImportJobPolicy()}
	exports := []iamTransferProcessorStub{
		{policy: iammodule.SystemExportJobPolicy()},
		{policy: iammodule.SystemExportReconcileJobPolicy()},
	}
	processors, err := appendIAMTransferProcessors(nil, imports,
		transferProcessors(exports[0], exports[1]))
	if err != nil || len(processors) != 3 {
		t.Fatalf("processors=%d err=%v", len(processors), err)
	}
	catalog, err := jobexecution.NewPolicyCatalog(iammodule.SystemTransferJobPolicies())
	if err != nil {
		t.Fatal(err)
	}
	if _, err = jobexecution.NewRegistryForCatalog(catalog, processors); err != nil {
		t.Fatalf("production transfer registry: %v", err)
	}
	if _, err = appendIAMTransferProcessors(nil, imports,
		transferProcessors(exports[0])); err == nil {
		t.Fatal("missing reconciliation processor was accepted")
	}
}

func transferProcessors(values ...iamTransferProcessorStub) []jobexecution.Processor {
	result := make([]jobexecution.Processor, len(values))
	for index := range values {
		result[index] = values[index]
	}
	return result
}
