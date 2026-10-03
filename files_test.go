package main

import (
	"path/filepath"
	"testing"

	"server/httpresponse"
)

func TestResolveFileTraversal(t *testing.T) {
	root, err := filepath.Abs("www")
	if err != nil {
		t.Fatal(err)
	}

	cases := []string{
		"/../../Windows/System32/drivers/etc/hosts",
		"/foo/../../../etc/hosts",
		"/../../../etc/passwd",
	}
	for _, path := range cases {
		if _, status := resolveFile(root, path); status != httpresponse.Forbidden {
			t.Fatalf("path %q: got %v, want 403", path, status)
		}
	}
}

func TestResolveFileIndexAndMissing(t *testing.T) {
	root, err := filepath.Abs("www")
	if err != nil {
		t.Fatal(err)
	}

	path, status := resolveFile(root, "/")
	if status != httpresponse.OK {
		t.Fatalf("GET / : %v", status)
	}
	if filepath.Base(path) != "index.html" {
		t.Fatalf("GET / resolved to %s", path)
	}

	if _, status := resolveFile(root, "/nao-existe.html"); status != httpresponse.NotFound {
		t.Fatalf("missing file: %v", status)
	}
}
