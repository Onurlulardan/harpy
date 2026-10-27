package postgresv1

import "testing"

func TestPostgresArtifactSchemasCompile(t *testing.T) {
	for _, name := range postgresSchemaNames {
		t.Run(name, func(t *testing.T) {
			compileArtifactSchema(t, name)
		})
	}
}
