package main

import (
	"path"
	"strings"
)

func importViolation(source guardedSource, importPath string) string {
	if message := removedTree(importPath); message != "" {
		return message
	}
	if violation := objectSDKViolation(importPath); violation != "" {
		return violation
	}
	switch source.kind {
	case "cmd":
		return commandViolation(source, importPath)
	case "setup":
		if violation := setupViolation(importPath); violation != "" {
			return violation
		}
	case "platform":
		if violation := platformViolation(importPath); violation != "" {
			return violation
		}
	case "module":
		return moduleViolation(source, importPath)
	}
	return awsGate(source, importPath)
}

func commandViolation(source guardedSource, importPath string) string {
	if stdlibImport(importPath) || importPath == compositionEntry(source.prefix, source.profile) {
		return ""
	}
	return "command must import only its matching composition entry"
}

func compositionEntry(prefix, profile string) string {
	root := modulePath
	if prefix != "" {
		root += "/" + prefix
	}
	return root + "/internal/composition/" + profile
}

func setupViolation(importPath string) string {
	_, layer, ok := moduleRef(importPath)
	if ok && (layer == "application" || layer == "adapter") {
		return "setup cannot import module application or adapter packages"
	}
	return ""
}

func platformViolation(importPath string) string {
	if _, _, ok := moduleRef(importPath); ok {
		return "platform cannot import modules"
	}
	return ""
}

func moduleViolation(source guardedSource, importPath string) string {
	if source.layer == "domain" {
		if violation := domainViolation(importPath); violation != "" {
			return violation
		}
	}
	if violation := moduleEdgeViolation(source, importPath); violation != "" {
		return violation
	}
	return awsGate(source, importPath)
}

func moduleEdgeViolation(source guardedSource, importPath string) string {
	name, layer, ok := moduleRef(importPath)
	if !ok {
		return ""
	}
	if name != source.module {
		if layer == "api" {
			return ""
		}
		return "cross-module import must target the other module api package"
	}
	if sameModuleAllowed(source.layer, layer) {
		return ""
	}
	return "same-module dependency direction is forbidden"
}

func sameModuleAllowed(sourceLayer, targetLayer string) bool {
	switch sourceLayer {
	case "application":
		return targetLayer != "adapter"
	case "domain":
		return targetLayer != "adapter" && targetLayer != "application"
	default:
		return true
	}
}

func stdlibImport(importPath string) bool {
	if importPath == "C" || importPath == "" || strings.HasPrefix(importPath, ".") {
		return false
	}
	first, _, _ := strings.Cut(importPath, "/")
	if first == "" || strings.Contains(first, ".") {
		return false
	}
	return path.Clean(importPath) == importPath
}
