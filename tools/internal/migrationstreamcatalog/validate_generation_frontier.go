package migrationstreamcatalog

import "fmt"

func validateGenerationFrontier(streams []Stream) error {
	var platformFrontier int64
	var tenantFrontier int64
	for _, stream := range streams {
		switch stream.Scope {
		case "platform":
			platformFrontier = max(platformFrontier, stream.TargetDatabaseGeneration)
		case "tenant":
			tenantFrontier = max(tenantFrontier, stream.TargetDatabaseGeneration)
		}
	}
	if tenantFrontier > platformFrontier {
		return fmt.Errorf(
			"tenant migration generation frontier %d exceeds platform frontier %d",
			tenantFrontier,
			platformFrontier,
		)
	}
	return nil
}
