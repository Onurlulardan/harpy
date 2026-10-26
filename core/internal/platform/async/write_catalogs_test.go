package async

import (
	"testing"

	asyncmanifestv1 "github.com/Onurlulardan/harpy/contracts/asyncmanifest/v1"
)

func TestWriteCatalogsAreReleaseSealed(t *testing.T) {
	if _, _, err := (WriteCatalogs{}).Snapshot(); err == nil {
		t.Fatal("zero-value write catalogs were accepted")
	}
	catalogs, err := LoadWriteCatalogs()
	if err != nil {
		t.Fatal(err)
	}
	jobs, producers, err := catalogs.Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	manifest, err := asyncmanifestv1.Load()
	if err != nil {
		t.Fatal(err)
	}
	jobSHA, _ := jobs.SHA256()
	wantJobSHA, _ := manifest.WritableJobs().SHA256()
	producerSHA, _ := producers.SHA256()
	wantProducerSHA, _ := manifest.Producers().SHA256()
	if jobSHA != wantJobSHA || producerSHA != wantProducerSHA {
		t.Fatalf("release catalog drift = jobs %s/%s producers %s/%s",
			jobSHA, wantJobSHA, producerSHA, wantProducerSHA)
	}
}
