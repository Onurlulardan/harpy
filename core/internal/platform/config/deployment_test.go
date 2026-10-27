package config

import (
	"strings"
	"testing"
)

func TestLoadDeploymentModeDefaultsMissingToOnPremise(t *testing.T) {
	mode, err := loadDeploymentMode(environmentValues{})
	if err != nil || mode != DeploymentOnPrem {
		t.Fatalf("deployment mode = %q, %v", mode, err)
	}
}

func TestLoadDeploymentModeAcceptsExactValues(t *testing.T) {
	for _, expected := range []DeploymentMode{DeploymentOnPrem, DeploymentSaaS} {
		mode, err := loadDeploymentMode(environmentValues{
			deploymentModeEnvironmentKey: string(expected),
		})
		if err != nil || mode != expected {
			t.Fatalf("deployment mode = %q, %v; want %q", mode, err, expected)
		}
	}
}

func TestLoadDeploymentModeRejectsNonExactValues(t *testing.T) {
	for _, value := range []string{"", " ", "hosted", "SAAS", "OnPremise"} {
		_, err := loadDeploymentMode(environmentValues{deploymentModeEnvironmentKey: value})
		if err == nil || !strings.Contains(err.Error(), deploymentModeEnvironmentKey) {
			t.Fatalf("deployment mode %q error = %v", value, err)
		}
	}
}

func TestLoadWorkerDefaultsMissingDeploymentModeToOnPremise(t *testing.T) {
	loaded, err := loadWorker(validWorkerEnvironment())
	if err != nil {
		t.Fatal(err)
	}
	mode, err := loaded.DeploymentMode()
	if err != nil || mode != DeploymentOnPrem {
		t.Fatalf("worker deployment mode = %q, %v", mode, err)
	}
}
