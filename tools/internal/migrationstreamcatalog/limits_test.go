package migrationstreamcatalog

import (
	"fmt"
	"sort"
	"strings"
	"testing"
)

func TestLoadEnforcesRootAndFragmentByteLimits(t *testing.T) {
	t.Run("root", func(t *testing.T) {
		root := t.TempDir()
		writeRawFile(t, root, rootDocumentPath, []byte(strings.Repeat(" ", maximumRootBytes+1)))
		_, err := Load(root)
		if err == nil || !strings.Contains(err.Error(), "bounded regular file") {
			t.Fatalf("root limit error = %v", err)
		}
	})
	t.Run("fragment", func(t *testing.T) {
		root := t.TempDir()
		include := streamPath("iam")
		writeRoot(t, root, catalogVersion, []string{include})
		writeRawFile(t, root, include, []byte(strings.Repeat(" ", maximumFragmentBytes+1)))
		_, err := Load(root)
		if err == nil || !strings.Contains(err.Error(), "bounded regular file") {
			t.Fatalf("fragment limit error = %v", err)
		}
	})
}

func TestLoadEnforcesIncludeLimit(t *testing.T) {
	root := t.TempDir()
	includes := make([]string, maximumIncludes+1)
	for index := range includes {
		includes[index] = streamPath(fmt.Sprintf("owner_a%03d", index))
	}
	sort.Strings(includes)
	writeRoot(t, root, catalogVersion, includes)
	_, err := Load(root)
	if err == nil || !strings.Contains(err.Error(), "include limit") {
		t.Fatalf("include limit error = %v", err)
	}
}

func TestLoadEnforcesAggregateByteLimit(t *testing.T) {
	root := t.TempDir()
	includes := make([]string, 0, 16)
	for letter := 'a'; letter <= 'p'; letter++ {
		owner := "owner_" + string(letter)
		include := streamPath(owner)
		includes = append(includes, include)
		body := ownerBody(
			ownerDocumentVersion, owner, []Stream{validStream(owner, int64(letter))},
		)(t)
		body = append(body, []byte(strings.Repeat(" ", maximumFragmentBytes-len(body)))...)
		writeRawFile(t, root, include, body)
	}
	writeRoot(t, root, catalogVersion, includes)
	_, err := Load(root)
	if err == nil || !strings.Contains(err.Error(), "total byte limit") {
		t.Fatalf("aggregate limit error = %v", err)
	}
}
