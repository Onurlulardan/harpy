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
	parts := strings.Split(path, "/")
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

func segment(value string) bool {
	return value != "" && value != "." && value != ".." && !strings.ContainsAny(value, `\/`)
}
