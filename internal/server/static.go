package server

import (
	"embed"
	"io/fs"
	"net/http"
	"strings"
)

//go:embed web_dist/*
var embeddedWeb embed.FS

// FileSystemCustom returns the embedded filesystem or nil if not compiled.
func getEmbeddedFS() http.FileSystem {
	sub, err := fs.Sub(embeddedWeb, "web_dist")
	if err != nil {
		return nil
	}
	return http.FS(sub)
}

// ServeStaticSPA handles SPA routing: serves file if exists, otherwise serves index.html.
func ServeStaticSPA(w http.ResponseWriter, r *http.Request) {
	efs := getEmbeddedFS()
	if efs == nil {
		http.Error(w, "ClickLens Web Studio bundle is not compiled. Run 'npm run build' inside web/ and recompile binary.", http.StatusNotFound)
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/")
	if path == "" {
		path = "index.html"
	}

	f, err := efs.Open(path)
	if err != nil {
		// Route not found as asset, fallback to index.html for SPA client-side routing
		indexFile, errIndex := efs.Open("index.html")
		if errIndex != nil {
			http.Error(w, "ClickLens Web Studio index.html not found.", http.StatusNotFound)
			return
		}
		defer indexFile.Close()
		stat, _ := indexFile.Stat()
		http.ServeContent(w, r, "index.html", stat.ModTime(), indexFile)
		return
	}
	defer f.Close()

	stat, _ := f.Stat()
	if stat.IsDir() {
		indexFile, errIndex := efs.Open("index.html")
		if errIndex != nil {
			http.Error(w, "ClickLens Web Studio index.html not found.", http.StatusNotFound)
			return
		}
		defer indexFile.Close()
		statIdx, _ := indexFile.Stat()
		http.ServeContent(w, r, "index.html", statIdx.ModTime(), indexFile)
		return
	}

	http.ServeContent(w, r, stat.Name(), stat.ModTime(), f)
}
