package main

import "fmt"

type fullRolloutInvocation struct {
	ClusterAdminDSNFile string   `json:"cluster_admin_dsn_file"`
	CompletionEvidence  string   `json:"completion_evidence"`
	EvidenceArtifacts   []string `json:"evidence_artifacts"`
	authorityArtifactPaths
}

func (value fullRolloutInvocation) validate() error {
	if len(value.EvidenceArtifacts) == 0 || len(value.EvidenceArtifacts) > 128 {
		return fmt.Errorf("full rollout evidence artifact count is invalid")
	}
	paths := value.authorityArtifactPaths.all()
	paths = append(paths, value.ClusterAdminDSNFile, value.CompletionEvidence)
	paths = append(paths, value.EvidenceArtifacts...)
	return validateAbsoluteDistinctPaths(paths)
}
