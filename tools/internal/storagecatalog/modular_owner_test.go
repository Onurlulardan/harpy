package storagecatalog

import (
	"reflect"
	"strings"
	"testing"
)

func TestLoadMergesExplicitModularOwnerFragments(t *testing.T) {
	root := t.TempDir()
	owner := validOwner("platform_async")
	owner.PlatformRoutines = nil
	owner.ExternalAccesses = nil
	includes := []string{
		fragmentPath("platform_async", "01_platform_routines"),
		fragmentPath("platform_async", "02_external_accesses"),
	}
	writeRoot(t, root, catalogVersion, []string{ownerPath(owner.Name)})
	writeJSONFile(t, root, ownerPath(owner.Name), modularOwnerDocument{
		Version: modularOwnerVersion, Owner: owner, Includes: includes,
	})
	writeOwnerFragment(t, root, includes[0], ownerFragmentDocument{
		Version: ownerFragmentVersion, Owner: owner.Name,
		PlatformRoutines: []string{"platform.platform_async_run_v1"},
	})
	writeOwnerFragment(t, root, includes[1], ownerFragmentDocument{
		Version: ownerFragmentVersion, Owner: owner.Name,
		ExternalAccesses: []ExternalAccess{{
			Access: "read", TargetOwner: "tenancy",
			Target: "platform.tenants", Contract: "ADR-1",
		}},
	})

	manifest, err := Load(root)
	if err != nil {
		t.Fatal(err)
	}
	want := owner
	want.PlatformRoutines = []string{"platform.platform_async_run_v1"}
	want.ExternalAccesses = []ExternalAccess{{
		Access: "read", TargetOwner: "tenancy",
		Target: "platform.tenants", Contract: "ADR-1",
	}}
	if !reflect.DeepEqual(manifest.Owners, []Owner{want}) {
		t.Fatalf("owners = %#v, want %#v", manifest.Owners, []Owner{want})
	}
}

func TestLoadRejectsInvalidModularOwnerFragments(t *testing.T) {
	tests := []struct {
		name  string
		apply func(*testing.T, string, string)
		want  string
	}{
		{"wrong_owner", writeWrongFragmentOwner, "declares owner"},
		{"empty_fragment", writeEmptyFragment, "at least one owner array"},
		{"orphan_fragment", writeOrphanFragment, "orphan"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			root := t.TempDir()
			owner := validOwner("iam")
			owner.PlatformTables = nil
			include := fragmentPath(owner.Name, "01_platform_tables")
			writeRoot(t, root, catalogVersion, []string{ownerPath(owner.Name)})
			writeJSONFile(t, root, ownerPath(owner.Name), modularOwnerDocument{
				Version: modularOwnerVersion, Owner: owner, Includes: []string{include},
			})
			writeOwnerFragment(t, root, include, ownerFragmentDocument{
				Version: ownerFragmentVersion, Owner: owner.Name,
				PlatformTables: []string{"platform.iam_users"},
			})
			test.apply(t, root, include)
			_, err := Load(root)
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("Load error = %v, want %q", err, test.want)
			}
		})
	}
}

func writeWrongFragmentOwner(t *testing.T, root, include string) {
	writeOwnerFragment(t, root, include, ownerFragmentDocument{
		Version: ownerFragmentVersion, Owner: "other", PlatformTables: []string{},
	})
}

func writeEmptyFragment(t *testing.T, root, include string) {
	writeOwnerFragment(t, root, include, ownerFragmentDocument{
		Version: ownerFragmentVersion, Owner: "iam",
	})
}

func writeOrphanFragment(t *testing.T, root, _ string) {
	writeOwnerFragment(t, root, fragmentPath("iam", "02_orphan"), ownerFragmentDocument{
		Version: ownerFragmentVersion, Owner: "iam", PlatformTables: []string{},
	})
}

func writeOwnerFragment(
	t *testing.T, root, include string, document ownerFragmentDocument,
) {
	t.Helper()
	writeJSONFile(t, root, include, document)
}

func fragmentPath(owner, name string) string {
	return fragmentDirectoryPath + "/" + owner + "/" + name + ".json"
}
