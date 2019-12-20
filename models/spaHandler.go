package models

import (
	"net/http"
	"os"
	"path/filepath"
	// "fmt"
)

// SpaHandler model to serve static files for SPA
type SpaHandler struct {
	StaticPath string
	IndexPath  string
}

// ServeHTTP inspects the URL path to locate a file within the static dir
// on the SPA handler If a file is found, it will be served. If not, the
// file located at the index path on the SPA handler wii be served. This
// is suitable behaviour for serving an SPA.

func (h *SpaHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// get the absolute path to prevent directory traversal
	path, err := filepath.Abs(r.URL.Path)

	if err != nil {
		// if we failed to ge the absolute path respond with a 400 bad request and stop
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// prepend the path with the pth to the static directory
	path = filepath.Join(h.StaticPath, path)

	// check whether a file existis at the given path
	_, err = os.Stat(path)

	if os.IsNotExist(err) {
		// file does not exist, serve index.html

		http.ServeFile(w, r, filepath.Join(h.StaticPath, h.IndexPath))
		return
	} else if err != nil {
		// if we get an error (that wasn't that the file does't exist) stating the
		// file, return a 500 internal server error and stop
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// otherwise, use http.FileServer to serve the static dir
	http.FileServer(http.Dir(h.StaticPath)).ServeHTTP(w, r)
}
