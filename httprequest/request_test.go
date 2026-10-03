package httprequest

import (
	"bufio"
	"strings"
	"testing"
)

func TestReadValidGet(t *testing.T) {
	raw := "GET /notas.txt HTTP/1.1\r\nHost: localhost\r\n\r\n"
	req, err := Read(bufio.NewReader(strings.NewReader(raw)))
	if err != nil {
		t.Fatal(err)
	}
	if req.Method != "GET" || req.Path != "/notas.txt" || req.Close {
		t.Fatalf("unexpected request: %+v", req)
	}
}

func TestReadPercentEncodingAndLeftover(t *testing.T) {
	raw := "GET /%20docs%2Ffile.txt HTTP/1.1\r\nHost: localhost\r\n\r\nGET /next HTTP/1.1\r\nHost: localhost\r\n\r\n"
	reader := bufio.NewReader(strings.NewReader(raw))
	req, err := Read(reader)
	if err != nil {
		t.Fatal(err)
	}
	if req.Path != "/ docs/file.txt" {
		t.Fatalf("decoded path = %q", req.Path)
	}

	next, err := Read(reader)
	if err != nil {
		t.Fatal(err)
	}
	if next.Path != "/next" {
		t.Fatalf("leftover path = %q", next.Path)
	}
}

func TestReadBadRequestLineAndHeader(t *testing.T) {
	if _, err := Read(bufio.NewReader(strings.NewReader("FOO\r\n\r\n"))); err != ErrBadRequest {
		t.Fatalf("short request-line: %v", err)
	}
	if _, err := Read(bufio.NewReader(strings.NewReader("GET / HTTP/1.1\r\nQuebrado\r\n\r\n"))); err != ErrBadRequest {
		t.Fatalf("header without colon: %v", err)
	}
}

func TestConnectionClose(t *testing.T) {
	raw := "GET / HTTP/1.1\r\nHost: localhost\r\nConnection: close\r\n\r\n"
	req, err := Read(bufio.NewReader(strings.NewReader(raw)))
	if err != nil {
		t.Fatal(err)
	}
	if !req.Close {
		t.Fatal("expected Connection: close")
	}
}

func TestHTTP11RequiresHost(t *testing.T) {
	if _, err := Read(bufio.NewReader(strings.NewReader("GET / HTTP/1.1\r\n\r\n"))); err != ErrBadRequest {
		t.Fatalf("HTTP/1.1 without Host: %v", err)
	}
	if _, err := Read(bufio.NewReader(strings.NewReader("GET / HTTP/1.1\r\nHost: a\r\nHost: b\r\n\r\n"))); err != ErrBadRequest {
		t.Fatalf("duplicate Host: %v", err)
	}
	if _, err := Read(bufio.NewReader(strings.NewReader("GET / HTTP/1.1\r\nHost: a b\r\n\r\n"))); err != ErrBadRequest {
		t.Fatalf("invalid Host: %v", err)
	}
	req, err := Read(bufio.NewReader(strings.NewReader("GET / HTTP/1.0\r\n\r\n")))
	if err != nil {
		t.Fatal(err)
	}
	if req.Version != "HTTP/1.0" {
		t.Fatalf("HTTP/1.0 without Host should be accepted: %+v", req)
	}
}
