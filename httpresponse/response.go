package httpresponse

import "fmt"

func BuildResponse(status Status, body []byte, contentType string) string {
	response := fmt.Sprintf(
		"HTTP/1.1 %d %s\r\n"+
			"Content-Type: %s\r\n"+
			"Content-Length: %d\r\n"+
			"Connection: close\r\n"+
			"\r\n"+
			"%s",
		status.Code,
		status.Message,
		contentType,
		len(body),
		body,
	)
	return response
}
