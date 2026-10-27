// Command coreopenapi rewrites the core API aggregate without deleted fragments.
package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const openAPIPath = "contracts/http/core-api-v1.openapi.yaml"

func main() {
	root, err := os.Getwd()
	if err != nil {
		exit(err)
	}
	documentPath := filepath.Join(root, filepath.FromSlash(openAPIPath))
	data, err := os.ReadFile(documentPath)
	if err != nil {
		exit(err)
	}
	var document map[string]any
	if err = json.Unmarshal(data, &document); err != nil {
		exit(err)
	}
	paths, _ := document["paths"].(map[string]any)
	for path, value := range paths {
		if removedReference(value) {
			delete(paths, path)
		}
	}
	encoded, err := encodeOpenAPI(document)
	if err != nil {
		exit(err)
	}
	if err = os.WriteFile(documentPath, encoded, 0o644); err != nil {
		exit(err)
	}
	sum := sha256.Sum256(encoded)
	input := sha256.Sum256(data)
	if err = updateOpenAPIManifest(root, hex.EncodeToString(input[:]), hex.EncodeToString(sum[:])); err != nil {
		exit(err)
	}
}

func encodeOpenAPI(document map[string]any) ([]byte, error) {
	var buffer bytes.Buffer
	buffer.WriteString("{\n")
	for _, key := range []string{"openapi", "x-code-generated"} {
		encoded, err := json.Marshal(document[key])
		if err != nil {
			return nil, err
		}
		fmt.Fprintf(&buffer, "  %q: %s,\n", key, encoded)
		delete(document, key)
	}
	rest, err := json.MarshalIndent(document, "  ", "  ")
	if err != nil {
		return nil, err
	}
	inner := strings.TrimSuffix(strings.TrimPrefix(string(rest), "{"), "}")
	inner = strings.Trim(inner, "\n")
	if inner != "" {
		buffer.WriteString(inner)
		buffer.WriteByte('\n')
	}
	buffer.WriteString("}\n")
	return buffer.Bytes(), nil
}

func removedReference(value any) bool {
	item, ok := value.(map[string]any)
	if !ok {
		return false
	}
	reference, _ := item["$ref"].(string)
	base := filepath.Base(filepath.FromSlash(strings.SplitN(reference, "#", 2)[0]))
	return strings.HasPrefix(base, "forms-") || strings.HasPrefix(base, "workflow-")
}

func updateOpenAPIManifest(root, inputSHA, outputSHA string) error {
	path := filepath.Join(root, "contracts", "generated.yaml")
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	var manifest struct {
		Version int                 `json:"version"`
		Files   []map[string]string `json:"files"`
	}
	if err = json.Unmarshal(data, &manifest); err != nil {
		return err
	}
	for _, file := range manifest.Files {
		if file["path"] != openAPIPath {
			continue
		}
		file["generator"] = "go run ./tools/coreopenapi"
		file["input_sha256"] = inputSHA
		file["output_sha256"] = outputSHA
	}
	encoded, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(encoded, '\n'), 0o644)
}

func exit(err error) {
	os.Stderr.WriteString(err.Error() + "\n")
	os.Exit(1)
}
