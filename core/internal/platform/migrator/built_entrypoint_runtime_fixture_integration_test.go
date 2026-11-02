//go:build integration

package migrator

func appendBuiltEntrypointRuntimeFixture(
	target []string,
	fixture []string,
) []string {
	return append(target, fixture...)
}
