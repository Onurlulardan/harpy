package processcode

import "testing"

func TestCodesExposePublishedProcessABI(t *testing.T) {
	wants := []int{0, 2, 3, 4, 5, 10}
	values := []Code{Clean, InvalidStartup, DependencyPreflight, ServeFailure, DrainForced, DeployMigration}
	for index, value := range values {
		if value.Int() != wants[index] || !value.Valid() {
			t.Fatalf("code %d is not published ABI", value)
		}
	}
}
