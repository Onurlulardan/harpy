package main

import "strings"

var modulePath string

type guardedSource struct {
	path    string
	kind    string
	layer   string
	module  string
	profile string
	prefix  string
}

func classifySource(source string) guardedSource {
	parts := strings.Split(source, "/")
	prefix, rest := productPrefix(parts)
	guarded := guardedSource{path: source, prefix: prefix}
	if len(rest) >= 2 && rest[0] == "cmd" {
		switch rest[1] {
		case "api", "worker", "migrator", "operator":
			guarded.kind = "cmd"
			guarded.profile = rest[1]
			return guarded
		case "setup":
			guarded.kind = "setup"
			guarded.profile = "setup"
			return guarded
		}
	}
	if len(rest) >= 2 && rest[0] == "internal" && rest[1] == "platform" {
		guarded.kind = "platform"
		return guarded
	}
	if len(rest) >= 3 && rest[0] == "internal" && rest[1] == "modules" && segment(rest[2]) {
		guarded.kind = "module"
		guarded.module = rest[2]
		guarded.layer = moduleLayer(rest[3 : len(rest)-1])
		return guarded
	}
	return guarded
}

func productPrefix(parts []string) (string, []string) {
	if len(parts) > 1 && parts[0] == "core" {
		return "core", parts[1:]
	}
	return "", parts
}

func moduleLayer(parts []string) string {
	if len(parts) == 0 {
		return "module"
	}
	if parts[0] == "adapter" || parts[0] == "adapters" {
		return "adapter"
	}
	return parts[0]
}

func moduleRef(importPath string) (string, string, bool) {
	parts := strings.Split(importPath, "/")
	for index := 0; index+2 < len(parts); index++ {
		if parts[index] != "internal" || parts[index+1] != "modules" || !segment(parts[index+2]) {
			continue
		}
		return parts[index+2], moduleLayer(parts[index+3:]), true
	}
	return "", "", false
}

func removedTree(path string) string {
	normalized := strings.ReplaceAll(path, "\\", "/")
	if message := removedContract(normalized); message != "" {
		return message
	}
	base := normalized
	if slash := strings.LastIndex(normalized, "/"); slash >= 0 {
		base = normalized[slash+1:]
	}
	if strings.HasPrefix(base, "forms_") {
		return "forms_ filename prefix is forbidden"
	}
	switch base {
	case "ai-runtime.json", "mcp-runtime.json":
		return "ai-runtime.json and mcp-runtime.json are forbidden"
	}
	parts := strings.Split(normalized, "/")
	for index := 0; index+1 < len(parts); index++ {
		switch parts[index] {
		case "modules":
			switch parts[index+1] {
			case "forms", "workflow":
				return "modules/forms and modules/workflow are forbidden"
			case "ai", "mcp":
				return "modules/ai and modules/mcp are forbidden"
			}
		case "services":
			switch parts[index+1] {
			case "ai", "mcp":
				return "services/ai and services/mcp are forbidden"
			}
		}
	}
	return ""
}

func removedContract(path string) string {
	markers := []struct{ fragment, message string }{
		{"contracts/aiinference", "contracts/aiinference is forbidden"},
		{"contracts/aiprovidercontext", "contracts/aiprovidercontext is forbidden"},
		{"contracts/aicredential", "contracts/aicredential is forbidden"},
		{"contracts/http/workflow-", "workflow contract paths are forbidden"},
		{"contracts/authority/workflow_", "workflow contract paths are forbidden"},
		{"contracts/operations/workflow_", "workflow contract paths are forbidden"},
		{"contracts/events/v1/workflow-", "workflow contract paths are forbidden"},
		{"contracts/consumers/workflowv2", "workflow contract paths are forbidden"},
	}
	for _, marker := range markers {
		if strings.Contains(path, marker.fragment) {
			return marker.message
		}
	}
	return ""
}

func segment(value string) bool {
	return value != "" && value != "." && value != ".." && !strings.ContainsAny(value, `\/`)
}
