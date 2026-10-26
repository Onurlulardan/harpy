package async

import (
	"testing"

	asyncmanifestv1 "github.com/Onurlulardan/harpy/contracts/asyncmanifest/v1"
)

func TestReleaseCatalogsAreSealedAndExact(t *testing.T) {
	if _, _, err := (ReleaseCatalogs{}).Snapshot(); err == nil {
		t.Fatal("zero-value release catalogs were accepted")
	}
	if _, err := (ReleaseCatalogs{}).GapJobs(); err == nil {
		t.Fatal("zero-value release gap catalog was accepted")
	}
	catalogs, err := LoadReleaseCatalogs()
	if err != nil {
		t.Fatal(err)
	}
	jobs, events, err := catalogs.Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	manifest, err := asyncmanifestv1.Load()
	if err != nil {
		t.Fatal(err)
	}
	jobSHA, _ := jobs.SHA256()
	wantJobSHA, _ := manifest.Jobs().SHA256()
	producerSHA, _ := events.Producers().SHA256()
	wantProducerSHA, _ := manifest.Producers().SHA256()
	consumerSHA, _ := events.Consumers().SHA256()
	wantConsumerSHA, _ := manifest.Consumers().SHA256()
	if jobSHA != wantJobSHA || producerSHA != wantProducerSHA ||
		consumerSHA != wantConsumerSHA {
		t.Fatal("platform release catalogs drifted from canonical manifest")
	}
	gaps, err := catalogs.GapJobs()
	if err != nil || len(gaps) != len(manifest.GapJobs()) {
		t.Fatalf("platform gap catalog drifted: %d/%d %v", len(gaps), len(manifest.GapJobs()), err)
	}
}
