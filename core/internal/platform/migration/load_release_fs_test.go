package migration

import (
	"bytes"
	"os"
	"testing"
)

func TestFilesystemReleaseMatchesEmbed(t *testing.T) {
	embedded, err := LoadReleaseRegistry()
	if err != nil {
		t.Fatal(err)
	}
	fromDisk, err := LoadReleaseManifestFS(os.DirFS("release"), ".")
	if err != nil {
		t.Fatal(err)
	}
	embeddedDigests := embedded.Digests()
	diskDigests := fromDisk.Digests()
	if len(embeddedDigests) == 0 || len(embeddedDigests) != len(diskDigests) {
		t.Fatalf("digest counts embedded %d disk %d", len(embeddedDigests), len(diskDigests))
	}
	for index, digest := range embeddedDigests {
		if diskDigests[index] != digest {
			t.Fatalf("digest %d embedded %s disk %s", index, digest, diskDigests[index])
		}
		if !bytes.Equal(embedded.manifests[digest], fromDisk.manifests[digest]) {
			t.Fatalf("canonical bytes differ for %s", digest)
		}
	}
	for _, index := range []int{0, len(embeddedDigests) / 2, len(embeddedDigests) - 1} {
		digest := embeddedDigests[index]
		artifact, err := fromDisk.MigrationArtifact(digest)
		if err != nil {
			t.Fatal(err)
		}
		if err = embedded.Authorize(artifact); err != nil {
			t.Fatalf("embed authorize %s: %v", digest, err)
		}
		if err = fromDisk.AuthorizeDigest(digest); err != nil {
			t.Fatal(err)
		}
	}
	if len(embedded.documentContracts) != len(fromDisk.documentContracts) {
		t.Fatalf("document contracts embedded %d disk %d",
			len(embedded.documentContracts), len(fromDisk.documentContracts))
	}
	for path, contracts := range embedded.documentContracts {
		other := fromDisk.documentContracts[path]
		if len(contracts) != len(other) {
			t.Fatalf("contracts for %s differ", path)
		}
		for contract := range contracts {
			if _, ok := other[contract]; !ok {
				t.Fatalf("disk registry missing contract for %s", path)
			}
		}
	}
	missing := "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
	if err = embedded.AuthorizeDigest(missing); err == nil || fromDisk.AuthorizeDigest(missing) == nil {
		t.Fatal("unknown digest was authorized")
	}
}
