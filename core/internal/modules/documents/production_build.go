package documents

import (
	"fmt"
)

func buildProduction(resources configuredProductionResources,
	generations HTTPGenerationResolver, scopes HTTPScopeCoordinator, buildID string,
) (*Production, error) {
	ports := []any{resources.artifacts.objects, resources.artifacts.keyring,
		resources.artifacts.cipher, resources.billing.actor, resources.jobs,
		resources.effects, generations, scopes}
	for _, port := range ports {
		if nilHTTPPort(port) {
			return nil, fmt.Errorf("Documents production graph is incomplete")
		}
	}
	ids, clock := platformIDs{}, platformClock{}
	application, err := NewPlatformApplicationRuntime(PlatformHTTPRuntimeDependencies{
		Generations: generations, Scopes: scopes, IDs: ids, Clock: clock,
		Billing: resources.billing.actor, UploadJobs: resources.jobs,
		ArchiveJobs: resources.jobs, Effects: resources.effects,
		Keyring: resources.artifacts.keyring, Cipher: resources.artifacts.cipher,
		Objects: resources.artifacts.objects, BuildID: buildID,
	})
	if err != nil {
		return nil, fmt.Errorf("build Documents application runtime: %w", err)
	}
	knowledge, err := NewKnowledgeFileOwner(KnowledgeFileOwnerDependencies{
		Generations: generations, Scopes: scopes, IDs: ids, Clock: clock,
		Billing: resources.billing.actor, Jobs: resources.jobs,
		Effects: resources.effects, Keyring: resources.artifacts.keyring,
		Cipher: resources.artifacts.cipher, Objects: resources.artifacts.objects,
		StepUp: stepUpVerifier{},
	})
	if err != nil {
		return nil, fmt.Errorf("build Documents knowledge-file owner: %w", err)
	}
	if nilHTTPPort(knowledge) {
		return nil, fmt.Errorf("Documents knowledge-file owner is unavailable")
	}
	return newProduction(application, knowledge)
}
