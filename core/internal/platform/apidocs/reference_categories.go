package apidocs

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"github.com/Onurlulardan/harpy/contracts"
	scalargo "github.com/bdpiprava/scalar-go"
)

const categorySpecPrefix = "http/category-"

var categoryOrder = []string{
	"Auth", "Billing", "Documents", "Invitations", "Notifications",
	"Directory", "Permissions", "Platform", "Tenancy", "Profile", "Application",
}

var (
	categoryOnce      sync.Once
	categoryDocuments map[string][]byte
	categoryLoadErr   error
)

func referenceSources(origin string) ([]scalargo.DocumentSource, error) {
	documents, err := loadCategoryDocuments()
	if err != nil {
		return nil, err
	}
	sources := make([]scalargo.DocumentSource, 0, len(categoryOrder)+1)
	for _, tag := range categoryOrder {
		if _, found := documents[tag]; !found {
			continue
		}
		sources = append(sources, scalargo.DocumentSource{
			Title: tag, Slug: strings.ToLower(tag),
			URL:     origin + specPrefix + categorySpecName(tag),
			Default: tag == "Auth",
		})
	}
	sources = append(sources, scalargo.DocumentSource{
		Title: "Management", Slug: "management",
		URL: origin + specPrefix + "http/management-v1.openapi.yaml",
	})
	if len(sources) == 1 || sources[0].Title != "Auth" {
		return nil, fmt.Errorf("API reference categories are incomplete")
	}
	return sources, nil
}

func categoryDocument(tag string) ([]byte, bool, error) {
	documents, err := loadCategoryDocuments()
	if err != nil {
		return nil, false, err
	}
	document, found := documents[tag]
	return document, found, nil
}

func categoryTagFromName(name string) (string, bool) {
	if !strings.HasPrefix(name, categorySpecPrefix) ||
		!strings.HasSuffix(name, ".openapi.json") {
		return "", false
	}
	slug := strings.TrimSuffix(strings.TrimPrefix(name, categorySpecPrefix), ".openapi.json")
	for _, tag := range categoryOrder {
		if strings.ToLower(tag) == slug {
			return tag, true
		}
	}
	return "", false
}

func categorySpecName(tag string) string {
	return categorySpecPrefix + strings.ToLower(tag) + ".openapi.json"
}

func loadCategoryDocuments() (map[string][]byte, error) {
	categoryOnce.Do(func() {
		categoryDocuments, categoryLoadErr = buildCategoryDocuments()
	})
	return categoryDocuments, categoryLoadErr
}

func buildCategoryDocuments() (map[string][]byte, error) {
	raw, err := contracts.Files.ReadFile("http/core-api-v1.openapi.yaml")
	if err != nil {
		return nil, err
	}
	var document map[string]any
	if err = json.Unmarshal(raw, &document); err != nil {
		return nil, err
	}
	paths, _ := document["paths"].(map[string]any)
	grouped := map[string]map[string]any{}
	for path, item := range paths {
		tag, found, tagErr := itemTag(item)
		if tagErr != nil {
			return nil, tagErr
		}
		if !found {
			continue
		}
		if grouped[tag] == nil {
			grouped[tag] = map[string]any{}
		}
		grouped[tag][path] = item
	}
	documents := make(map[string][]byte, len(grouped))
	for tag, tagPaths := range grouped {
		copyDocument := map[string]any{}
		for key, value := range document {
			copyDocument[key] = value
		}
		copyDocument["paths"] = tagPaths
		info, _ := copyDocument["info"].(map[string]any)
		if info == nil {
			info = map[string]any{}
		}
		infoCopy := map[string]any{}
		for key, value := range info {
			infoCopy[key] = value
		}
		infoCopy["title"] = tag
		infoCopy["description"] = tag + " operations."
		copyDocument["info"] = infoCopy
		encoded, encodeErr := json.Marshal(copyDocument)
		if encodeErr != nil {
			return nil, encodeErr
		}
		documents[tag] = encoded
	}
	return documents, nil
}

func itemTag(item any) (string, bool, error) {
	object, ok := item.(map[string]any)
	if !ok {
		return "", false, nil
	}
	if ref, found := object["$ref"].(string); found {
		return referenceTag(ref)
	}
	tag, found := methodTag(object)
	return tag, found, nil
}
