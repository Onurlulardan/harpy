package migrationstreamcatalog

import "sort"

type Manifest struct {
	Version int
	Owners  []Owner
}

type Owner struct {
	Name    string
	Streams []Stream
}

type Stream struct {
	Owner                        string `json:"owner"`
	Directory                    string `json:"directory"`
	Scope                        string `json:"scope"`
	VersionFrom                  int64  `json:"version_from"`
	VersionTo                    int64  `json:"version_to"`
	Phase                        string `json:"phase"`
	TargetDatabaseGeneration     int64  `json:"target_database_generation"`
	NextMinimumRuntimeGeneration int64  `json:"next_minimum_runtime_generation"`
}

func (manifest Manifest) Streams() []Stream {
	count := 0
	for _, owner := range manifest.Owners {
		count += len(owner.Streams)
	}
	streams := make([]Stream, 0, count)
	for _, owner := range manifest.Owners {
		streams = append(streams, owner.Streams...)
	}
	sort.Slice(streams, func(left, right int) bool {
		return streams[left].VersionFrom < streams[right].VersionFrom
	})
	return streams
}
