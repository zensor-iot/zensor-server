// Package web embeds the web frontend build output.
package web

import (
	"embed"
	"io/fs"
	"mime"
	"net/http"
	"path"
	"strings"
)

//go:embed all:dist
var distFS embed.FS

// SPAHandler serves the Vite-built React SPA embedded in dist. Unknown non-asset paths fall
// back to index.html so client-side routing (react-router) handles them; unknown /assets/*
// paths 404 for real.
func SPAHandler() http.Handler {
	sub, err := fs.Sub(distFS, "dist")
	if err != nil {
		panic(err)
	}
	fileServer := http.FileServer(http.FS(sub))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		urlPath := r.URL.Path

		isAsset := strings.HasPrefix(urlPath, "/assets/")
		if isAsset {
			w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
			if ct := assetContentType(urlPath); ct != "" {
				w = &contentTypeResponseWriter{ResponseWriter: w, contentType: ct}
			}
		} else {
			w.Header().Set("Cache-Control", "no-cache")
		}

		if urlPath != "/" {
			fsPath := strings.TrimPrefix(urlPath, "/")
			if _, err := fs.Stat(sub, fsPath); err != nil {
				if isAsset {
					http.NotFound(w, r)
					return
				}
				r.URL.Path = "/"
			}
		}

		fileServer.ServeHTTP(w, r)
	})
}

func assetContentType(urlPath string) string {
	if !strings.HasPrefix(urlPath, "/assets/") {
		return ""
	}
	return mime.TypeByExtension(path.Ext(urlPath))
}

type contentTypeResponseWriter struct {
	http.ResponseWriter
	contentType string
}

func (w *contentTypeResponseWriter) WriteHeader(statusCode int) {
	if w.contentType != "" {
		w.ResponseWriter.Header().Set("Content-Type", w.contentType)
	}
	w.ResponseWriter.WriteHeader(statusCode)
}
