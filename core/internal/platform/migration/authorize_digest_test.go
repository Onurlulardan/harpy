package migration

import (
	"strings"
	"testing"
)

func TestAuthorizeDigestAcceptsOnlyEmbeddedManifest(t *testing.T) {
	registry, err := LoadReleaseRegistry()
	if err != nil {
		t.Fatalf("load registry: %v", err)
	}
	digests := registry.Digests()
	if len(digests) == 0 {
		t.Fatal("expected embedded migration manifest")
	}
	if err := registry.AuthorizeDigest(digests[0]); err != nil {
		t.Fatalf("authorize embedded digest: %v", err)
	}
	if err := registry.AuthorizeDigest(strings.Repeat("f", 64)); err == nil {
		t.Fatal("unembedded digest was authorized")
	}
	if err := (*ReleaseRegistry)(nil).AuthorizeDigest(digests[0]); err == nil {
		t.Fatal("nil registry authorized a digest")
	}
}
