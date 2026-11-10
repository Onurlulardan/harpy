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
