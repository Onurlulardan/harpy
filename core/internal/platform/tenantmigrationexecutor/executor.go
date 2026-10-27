package tenantmigrationexecutor

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

type Dependencies struct {
	Store    Store
	Migrator Migrator
	Release  Release
	NewID    func() (uuid.UUID, error)
	ReadyTTL time.Duration
	LeaseTTL time.Duration
}

type Executor struct {
	store    Store
	migrator Migrator
	release  Release
	newID    func() (uuid.UUID, error)
	readyTTL time.Duration
	leaseTTL time.Duration
}

func New(dependencies Dependencies) (*Executor, error) {
	if dependencies.Store == nil || dependencies.Migrator == nil ||
		dependencies.NewID == nil || dependencies.Release.ID == uuid.Nil ||
		dependencies.Release.RuntimeGeneration < 1 ||
		!validDigest(dependencies.Release.MigrationManifestSHA256) ||
		(dependencies.Release.DeploymentMode != "onpremise" &&
			dependencies.Release.DeploymentMode != "saas") {
		return nil, fmt.Errorf("tenant migration executor dependencies are invalid")
	}
	readyTTL := dependencies.ReadyTTL
	if readyTTL == 0 {
		readyTTL = ReadyTTL
	}
	leaseTTL := dependencies.LeaseTTL
	if leaseTTL == 0 {
		leaseTTL = LeaseTTL
	}
	if readyTTL < 15*time.Second || readyTTL > 120*time.Second ||
		leaseTTL < 30*time.Second || leaseTTL > 300*time.Second {
		return nil, fmt.Errorf("tenant migration executor lease contract is invalid")
	}
	return &Executor{store: dependencies.Store, migrator: dependencies.Migrator,
		release: dependencies.Release, newID: dependencies.NewID,
		readyTTL: readyTTL, leaseTTL: leaseTTL}, nil
}

func validDigest(value string) bool {
	if len(value) != 64 {
		return false
	}
	for _, character := range value {
		if !((character >= '0' && character <= '9') ||
			(character >= 'a' && character <= 'f')) {
			return false
		}
	}
	return true
}
