package httpedge

import (
	"sync"
	"testing"
)

func TestAdmissionStartsClosedAndDrainIsIrreversible(t *testing.T) {
	admission := &Admission{}
	if admission.Accepting() {
		t.Fatal("admission started open")
	}
	if !admission.Open() || !admission.Accepting() {
		t.Fatal("admission did not open")
	}
	admission.Pause()
	if admission.Accepting() {
		t.Fatal("admission did not pause")
	}
	admission.BeginDrain()
	if !admission.Draining() || admission.Open() || admission.Accepting() {
		t.Fatal("draining admission reopened")
	}
}

func TestAdmissionConcurrentDrainWinsOverOpen(t *testing.T) {
	admission := &Admission{}
	var wait sync.WaitGroup
	for index := 0; index < 50; index++ {
		wait.Add(1)
		go func() {
			defer wait.Done()
			for attempt := 0; attempt < 100; attempt++ {
				admission.Open()
			}
		}()
	}
	admission.BeginDrain()
	wait.Wait()
	if !admission.Draining() || admission.Accepting() {
		t.Fatal("concurrent open defeated drain")
	}
}
