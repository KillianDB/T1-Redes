package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"

	"server/httprequest"
	"server/httpresponse"
)

const idleTimeout = 5 * time.Second

func main() {
	port := flag.Int("port", 0, "porta TCP do servidor (acima de 1024)")
	root := flag.String("root", "", "diretorio raiz a ser servido")
	flag.Parse()

	if *port <= 1024 || *port > 65535 {
		fmt.Fprintln(os.Stderr, "uso: --port <porta> --root <diretorio>")
		fmt.Fprintln(os.Stderr, "a porta deve ser um numero entre 1025 e 65535")
		os.Exit(1)
	}
	if *root == "" {
		fmt.Fprintln(os.Stderr, "uso: --port <porta> --root <diretorio>")
		os.Exit(1)
	}

	info, err := os.Stat(*root)
	if err != nil || !info.IsDir() {
		fmt.Fprintf(os.Stderr, "diretorio raiz invalido: %s\n", *root)
		os.Exit(1)
	}

	rootAbs, err := filepath.Abs(*root)
	if err != nil {
		fmt.Fprintf(os.Stderr, "nao foi possivel resolver o diretorio raiz: %v\n", err)
		os.Exit(1)
	}

	addr := fmt.Sprintf("0.0.0.0:%d", *port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "falha ao escutar em %s: %v\n", addr, err)
		os.Exit(1)
	}
	defer listener.Close()
	fmt.Printf("servidor escutando em %s, root=%s\n", addr, rootAbs)

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("falha ao aceitar conexao:", err)
			continue
		}
		go handleConnection(conn, rootAbs)
	}
}

func handleConnection(conn net.Conn, root string) {
	defer conn.Close()
	reader := bufio.NewReader(conn)
	remote := conn.RemoteAddr().String()
	fmt.Println(time.Now().Format(time.RFC3339), "conexao aberta", remote)

	for {
		if err := conn.SetReadDeadline(time.Now().Add(idleTimeout)); err != nil {
			return
		}

		// O bufio.Reader acumula o fluxo TCP e guarda bytes da proxima
		// requisicao, se uma leitura trouxer mais de uma mensagem.
		req, err := httprequest.Read(reader)
		if err != nil {
			if errors.Is(err, httprequest.ErrBadRequest) {
				_ = writeStatus(conn, httpresponse.BadRequest, true, true)
			} else if isIdleOrClosed(err) {
				fmt.Println(time.Now().Format(time.RFC3339), "conexao encerrada", remote)
			} else {
				fmt.Println(time.Now().Format(time.RFC3339), remote, "leitura:", err)
			}
			return
		}

		resp := dispatch(req, root)
		resp.Close = req.Close || resp.Close
		logRequest(remote, req, resp.Status)

		if err := httpresponse.Write(conn, resp); err != nil {
			fmt.Println(time.Now().Format(time.RFC3339), remote, "escrita:", err)
			return
		}
		if resp.Close {
			return
		}
	}
}

func dispatch(req *httprequest.Request, root string) httpresponse.Response {
	method := strings.ToUpper(req.Method)
	includeBody := method != "HEAD"

	if method != "GET" && method != "HEAD" {
		return errorResponse(httpresponse.MethodNotAllowed, includeBody, req.Close, "GET, HEAD")
	}

	path, status := resolveFile(root, req.Path)
	if status != httpresponse.OK {
		return errorResponse(status, includeBody, req.Close, "")
	}

	body, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return errorResponse(httpresponse.NotFound, includeBody, req.Close, "")
		}
		return errorResponse(httpresponse.InternalServerError, includeBody, req.Close, "")
	}

	return httpresponse.Response{
		Status:      httpresponse.OK,
		Body:        body,
		ContentType: contentType(path),
		Close:       req.Close,
		IncludeBody: includeBody,
	}
}

func errorResponse(status httpresponse.Status, includeBody, closeConn bool, allow string) httpresponse.Response {
	return httpresponse.Response{
		Status:      status,
		Body:        httpresponse.ErrorBody(status),
		ContentType: "text/plain",
		Close:       closeConn || status.Code == 400,
		Allow:       allow,
		IncludeBody: includeBody,
	}
}

func writeStatus(w io.Writer, status httpresponse.Status, includeBody, closeConn bool) error {
	return httpresponse.Write(w, errorResponse(status, includeBody, closeConn, ""))
}

func logRequest(remote string, req *httprequest.Request, status httpresponse.Status) {
	fmt.Printf("%s %s %s %s -> %d %s\n",
		time.Now().Format(time.RFC3339),
		remote,
		req.Method,
		req.Target,
		status.Code,
		status.Message,
	)
}

func isIdleOrClosed(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, io.EOF) {
		return true
	}
	var netErr net.Error
	return errors.As(err, &netErr) && netErr.Timeout()
}
