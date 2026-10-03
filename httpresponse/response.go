package httpresponse

import (
	"fmt"
	"io"
	"strings"
	"time"
)

const ServerName = "T1-Redes-Grupo"

type Response struct {
	Status      Status
	Body        []byte
	ContentType string
	Close       bool
	Allow       string
	IncludeBody bool
}

func imfFixdate(t time.Time) string {
	return t.UTC().Format("Mon, 02 Jan 2006 15:04:05 GMT")
}

func BuildHeaders(resp Response) string {
	var b strings.Builder
	fmt.Fprintf(&b, "HTTP/1.1 %d %s\r\n", resp.Status.Code, resp.Status.Message)
	fmt.Fprintf(&b, "Content-Type: %s\r\n", resp.ContentType)
	fmt.Fprintf(&b, "Content-Length: %d\r\n", len(resp.Body))
	fmt.Fprintf(&b, "Date: %s\r\n", imfFixdate(time.Now()))
	fmt.Fprintf(&b, "Server: %s\r\n", ServerName)
	if resp.Close {
		b.WriteString("Connection: close\r\n")
	} else {
		b.WriteString("Connection: keep-alive\r\n")
	}
	if resp.Allow != "" {
		fmt.Fprintf(&b, "Allow: %s\r\n", resp.Allow)
	}
	b.WriteString("\r\n")
	return b.String()
}

func Write(w io.Writer, resp Response) error {
	if _, err := io.WriteString(w, BuildHeaders(resp)); err != nil {
		return err
	}
	if resp.IncludeBody && len(resp.Body) > 0 {
		_, err := w.Write(resp.Body)
		return err
	}
	return nil
}

func ErrorBody(status Status) []byte {
	return []byte(fmt.Sprintf("%d %s\n", status.Code, status.Message))
}
