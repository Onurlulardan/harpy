package main

import (
	"bufio"
	"bytes"
	"os"
	"strings"
	"unicode/utf8"
)

const maximumTextLineBytes = 65 * 1024 * 1024

func isTextFile(path string) (bool, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return false, err
	}
	return bytes.IndexByte(content, 0) < 0 && utf8.Valid(content), nil
}

func countLines(path string) (int, error) {
	file, err := os.Open(path)
	if err != nil {
		return 0, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 64*1024), maximumTextLineBytes)
	count := 0
	for scanner.Scan() {
		count++
	}
	return count, scanner.Err()
}

func hasGeneratedHeader(path string) (bool, error) {
	file, err := os.Open(path)
	if err != nil {
		return false, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 64*1024), maximumTextLineBytes)
	for line := 0; line < 5 && scanner.Scan(); line++ {
		text := strings.ToLower(scanner.Text())
		if strings.Contains(text, "code generated") && strings.Contains(text, "do not edit") {
			return true, nil
		}
	}
	return false, scanner.Err()
}

func isLockfile(name string) bool {
	return name == "go.sum" || name == "yarn.lock" ||
		name == "package-lock.json" || name == "pnpm-lock.yaml" ||
		strings.HasSuffix(name, ".lock")
}

func isGeneratedBinary(name string) bool {
	return strings.HasSuffix(name, ".zip")
}
