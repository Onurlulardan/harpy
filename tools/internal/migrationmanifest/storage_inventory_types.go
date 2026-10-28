package migrationmanifest

import "github.com/Onurlulardan/harpy/tools/internal/migrationstreamcatalog"

type StorageInventory struct {
	claims  []StorageClaim
	streams []MigrationStream
}

type MigrationStream = migrationstreamcatalog.Stream

type StorageClaim struct {
	Owner     string
	Directory string
	Scope     string
}

func (inventory StorageInventory) Claims() []StorageClaim {
	return append([]StorageClaim(nil), inventory.claims...)
}

func (inventory StorageInventory) Streams() []MigrationStream {
	return append([]MigrationStream(nil), inventory.streams...)
}
