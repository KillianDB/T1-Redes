package main

import (
	"os"
	"path/filepath"
	"strings"

	"server/httpresponse"
)

func contentType(path string) string {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".html", ".htm":
		return "text/html"
	case ".css":
		return "text/css"
	case ".js":
		return "application/javascript"
	case ".json":
		return "application/json"
	case ".txt":
		return "text/plain"
	case ".gif":
		return "image/gif"
	case ".png":
		return "image/png"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".pdf":
		return "application/pdf"
	case ".mp4":
		return "video/mp4"
	default:
		return "application/octet-stream"
	}
}

func resolveFile(root, reqPath string) (string, httpresponse.Status) {
	rootAbs, err := filepath.Abs(root)
	if err != nil {
		return "", httpresponse.InternalServerError
	}

	if reqPath == "*" {
		return "", httpresponse.NotFound
	}

	rel := strings.TrimPrefix(reqPath, "/")
	if reqPath == "/" || strings.HasSuffix(reqPath, "/") {
		rel = filepath.Join(rel, "index.html")
	}

	candidate := filepath.Clean(filepath.Join(rootAbs, filepath.FromSlash(rel)))
	if !insideRoot(rootAbs, candidate) {
		return "", httpresponse.Forbidden
	}

	info, err := os.Stat(candidate)
	if err != nil {
		return "", httpresponse.NotFound
	}
	if info.IsDir() {
		candidate = filepath.Join(candidate, "index.html")
		if !insideRoot(rootAbs, candidate) {
			return "", httpresponse.Forbidden
		}
		info, err = os.Stat(candidate)
		if err != nil || info.IsDir() {
			return "", httpresponse.NotFound
		}
	}

	return candidate, httpresponse.OK
}

func insideRoot(root, candidate string) bool {
	rel, err := filepath.Rel(root, candidate)
	if err != nil {
		return false
	}
	return rel != ".." && !strings.HasPrefix(rel, ".."+string(os.PathSeparator))
}
