package apidocs

import (
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/Onurlulardan/harpy/contracts"
)

type specHandler struct{}

func (specHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimPrefix(r.URL.Path, specPrefix)
	if tag, found := categoryTagFromName(name); found {
		writeCategorySpec(w, r, tag)
		return
	}
	if !specName(name) {
		http.NotFound(w, r)
		return
	}
	file, err := contracts.Files.Open(name)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil || info.IsDir() {
		http.NotFound(w, r)
		return
	}
	seeker, ok := file.(io.ReadSeeker)
	if !ok {
		http.NotFound(w, r)
		return
	}
	http.ServeContent(w, r, info.Name(), time.Time{}, seeker)
}

func writeCategorySpec(w http.ResponseWriter, r *http.Request, tag string) {
	document, found, err := categoryDocument(tag)
	if err != nil || !found {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(document)
}

func specName(name string) bool {
	if name == "" || strings.HasPrefix(name, "/") || strings.Contains(name, "\\") {
		return false
	}
	for _, segment := range strings.Split(name, "/") {
		if segment == "" || segment == "." || segment == ".." {
			return false
		}
	}
	return true
}
