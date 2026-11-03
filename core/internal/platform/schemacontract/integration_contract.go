//go:build integration

package schemacontract

import "fmt"

func IntegrationContract(extensionManifestSHA256 string) (EmbeddedContract, error) {
	if extensionManifestSHA256 == "" {
		return EmbeddedContract{}, fmt.Errorf("extension manifest digest is empty")
	}
	return LoadBoundStartup()
}
