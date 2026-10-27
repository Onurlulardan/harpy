//go:build integration

package integrationfixture

import (
	"embed"
	"fmt"
)

//go:embed testdata/catalog_demo/*.fixture
var catalogDemoFiles embed.FS

func catalogDemoFixture(name string) []byte {
	document, err := catalogDemoFiles.ReadFile("testdata/catalog_demo/" + name)
	if err != nil {
		panic(fmt.Sprintf("read Catalog demo integration fixture %s: %v", name, err))
	}
	return document
}
