package httprequest

import (
	"bufio"
	"errors"
	"net/url"
	"strings"
)

const maxHeaderBytes = 65536

var (
	ErrBadRequest     = errors.New("bad request")
	ErrHeaderTooLarge = errors.New("header too large")
)

type Request struct {
	Method  string
	Target  string
	Version string
	Path    string
	Headers map[string]string
	Close   bool
}

func Read(reader *bufio.Reader) (*Request, error) {
	raw, err := readHeaderBlock(reader)
	if err != nil {
		return nil, err
	}

	text := strings.TrimSuffix(string(raw), "\r\n\r\n")
	lines := strings.Split(text, "\r\n")
	if len(lines) == 0 || lines[0] == "" {
		return nil, ErrBadRequest
	}

	method, target, version, err := parseRequestLine(lines[0])
	if err != nil {
		return nil, err
	}

	headers, err := parseHeaders(lines[1:])
	if err != nil {
		return nil, err
	}

	path, err := decodePath(target)
	if err != nil {
		return nil, err
	}

	if err := checkHost(version, headers); err != nil {
		return nil, err
	}

	req := &Request{
		Method:  method,
		Target:  target,
		Version: version,
		Path:    path,
		Headers: headers,
		Close:   wantsClose(version, headers),
	}
	return req, nil
}

func readHeaderBlock(reader *bufio.Reader) ([]byte, error) {
	var buf []byte
	for {
		b, err := reader.ReadByte()
		if err != nil {
			return nil, err
		}
		buf = append(buf, b)
		if len(buf) > maxHeaderBytes {
			return nil, ErrHeaderTooLarge
		}
		if len(buf) >= 4 && string(buf[len(buf)-4:]) == "\r\n\r\n" {
			return buf, nil
		}
	}
}

func parseRequestLine(line string) (method, target, version string, err error) {
	parts := strings.Split(line, " ")
	if len(parts) != 3 || parts[0] == "" || parts[1] == "" || parts[2] == "" {
		return "", "", "", ErrBadRequest
	}
	method, target, version = parts[0], parts[1], parts[2]
	if version != "HTTP/1.0" && version != "HTTP/1.1" {
		return "", "", "", ErrBadRequest
	}
	return method, target, version, nil
}

func parseHeaders(lines []string) (map[string]string, error) {
	headers := make(map[string]string)
	for _, line := range lines {
		if line == "" {
			return nil, ErrBadRequest
		}
		colon := strings.Index(line, ":")
		if colon < 0 {
			return nil, ErrBadRequest
		}
		name := strings.ToLower(strings.TrimSpace(line[:colon]))
		if name == "" {
			return nil, ErrBadRequest
		}
		value := strings.TrimSpace(line[colon+1:])
		if name == "host" {
			if _, exists := headers["host"]; exists {
				return nil, ErrBadRequest
			}
		}
		headers[name] = value
	}
	return headers, nil
}

func checkHost(version string, headers map[string]string) error {
	host, ok := headers["host"]
	if version == "HTTP/1.1" && !ok {
		return ErrBadRequest
	}
	if ok && !validHostValue(host) {
		return ErrBadRequest
	}
	return nil
}

func validHostValue(host string) bool {
	if host == "" {
		return true
	}
	if strings.ContainsAny(host, " \t@/") {
		return false
	}
	if strings.HasPrefix(host, "[") {
		end := strings.Index(host, "]")
		if end < 0 {
			return false
		}
		rest := host[end+1:]
		if rest == "" {
			return true
		}
		if !strings.HasPrefix(rest, ":") {
			return false
		}
		return isPort(rest[1:])
	}
	_, port, found := strings.Cut(host, ":")
	if found && !isPort(port) {
		return false
	}
	return true
}

func isPort(port string) bool {
	if port == "" {
		return false
	}
	for _, c := range port {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

func decodePath(target string) (string, error) {
	rawPath := target
	if i := strings.Index(rawPath, "#"); i >= 0 {
		rawPath = rawPath[:i]
	}

	if strings.Contains(rawPath, "://") {
		u, err := url.Parse(rawPath)
		if err != nil {
			return "", ErrBadRequest
		}
		rawPath = u.EscapedPath()
		if rawPath == "" {
			rawPath = "/"
		}
	} else {
		if i := strings.Index(rawPath, "?"); i >= 0 {
			rawPath = rawPath[:i]
		}
	}

	if rawPath == "" || (!strings.HasPrefix(rawPath, "/") && rawPath != "*") {
		return "", ErrBadRequest
	}

	decoded, err := url.PathUnescape(rawPath)
	if err != nil {
		return "", ErrBadRequest
	}
	return decoded, nil
}

func wantsClose(version string, headers map[string]string) bool {
	if conn, ok := headers["connection"]; ok {
		for _, part := range strings.Split(conn, ",") {
			token := strings.ToLower(strings.TrimSpace(part))
			if token == "close" {
				return true
			}
			if token == "keep-alive" {
				return false
			}
		}
	}
	return version != "HTTP/1.1"
}
