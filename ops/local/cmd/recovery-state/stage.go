package main

import (
	"fmt"
	"path/filepath"

	recoverykitv1 "github.com/Onurlulardan/harpy/contracts/recoverykit/v1"
)

func runStage(arguments []string, targetName string, validateKit bool) error {
	source, targetDirectory, err := parseStageFlags(arguments)
	if err != nil {
		return err
	}
	if !filepath.IsAbs(targetDirectory) {
		return fmt.Errorf("handoff directory is not absolute")
	}
	content, err := readRegular(source, maximumHandoffBytes)
	if err != nil {
		return err
	}
	if validateKit {
		if _, err := recoverykitv1.Parse(content); err != nil {
			return fmt.Errorf("recovery kit is invalid")
		}
	}
	return writeAtomic(filepath.Join(targetDirectory, targetName), content)
}
