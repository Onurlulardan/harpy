//go:build integration

package migrator

import "strings"

func appendBuiltEntrypointRuntimeFixture(
	target []string,
	fixture []string,
	catalogPackageKeyring string,
) []string {
	for _, value := range fixture {
		if catalogPackageKeyring != "" &&
			strings.HasPrefix(value, "HARPY_CATALOG_PACKAGE_PUBLIC_KEYRING=") {
			continue
		}
		target = append(target, value)
	}
	return target
}
