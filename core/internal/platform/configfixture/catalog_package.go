package configfixture

const catalogPackageKeyringEnvironmentKey = "HARPY_CATALOG_PACKAGE_PUBLIC_KEYRING"

func setCatalogPackageKeyring(setter EnvironmentSetter) {
	setter.Helper()
	setter.Setenv(catalogPackageKeyringEnvironmentKey, catalogPackageKeyring())
}

func catalogPackageKeyringEnvironment() string {
	return catalogPackageKeyringEnvironmentKey + "=" + catalogPackageKeyring()
}

func catalogPackageKeyring() string {
	return marshal(map[string]any{
		"version": 1,
		"keys": []any{map[string]string{
			"key_id": "release-2026-07", "public_key": material(7),
		}},
	})
}
