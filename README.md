# T1-Redes

Servidor HTTP/1.1 sobre sockets TCP, feito em Golang sem o uso de `net/http` do Golang. Trabalho universitário da cadeira de Laboratório de Redes - PUCRS.

**A pergunta central que o trabalho deve responder é:** como o comportamento do TCP determina o desempenho percebido do HTTP?

## Compilação

Na raiz do repositório, com o Go instalado:

```bash
go build -o servidor .
```

No Windows:

```powershell
go build -o servidor.exe .
```

## Execução

```bash
./servidor --port 8080 --root ./www
```

Argumentos obrigatórios:

- `--port` — porta TCP acima de 1024
- `--root` — diretório raiz dos arquivos a servir

O servidor escuta em `0.0.0.0` (todas as interfaces).

## Exemplos de teste

```bash
# 200 OK
curl -i http://<IP-servidor>:8080/

# HEAD (mesmos cabeçalhos do GET, sem corpo)
curl -i -I http://<IP-servidor>:8080/notas.txt

# 404 Not Found
curl -i http://<IP-servidor>:8080/nao-existe.html

# 405 Method Not Allowed
curl -i -X POST http://<IP-servidor>:8080/

# 400 Bad Request (request-line inválida)
printf 'FOO\r\n\r\n' | nc <IP-servidor> 8080

# 400 Bad Request (header sem ':')
printf 'GET / HTTP/1.1\r\nQuebrado\r\n\r\n' | nc <IP-servidor> 8080

# 400 Bad Request (HTTP/1.1 sem Host — RFC 9112 §3.2)
printf 'GET / HTTP/1.1\r\n\r\n' | nc <IP-servidor> 8080

# Mesmos 400 no PowerShell, sem nc:
$c = New-Object System.Net.Sockets.TcpClient("<IP-servidor>", 8080)
$s = $c.GetStream()
$b = [Text.Encoding]::ASCII.GetBytes("GET / HTTP/1.1`r`nQuebrado`r`n`r`n")
$s.Write($b, 0, $b.Length)

# extensao desconhecida -> application/octet-stream
curl -i http://<IP-servidor>:8080/arquivo.bin

# 403 Forbidden — travessia de diretório
curl -i --path-as-is http://<IP-servidor>:8080/../../Windows/System32/drivers/etc/hosts
curl -i --path-as-is http://<IP-servidor>:8080/foo/../../../etc/hosts
curl -i --path-as-is "http://<IP-servidor>:8080/%2e%2e/%2e%2e/%2e%2e/etc/passwd"

# persistência: fecha a conexão
curl -i -H "Connection: close" http://<IP-servidor>:8080/notas.txt
```

## Medições C1 e C2 (Parte 2)

Entre máquinas distintas, com Wireshark em `tcp.port == 8080` e o RTT do `ping` anotado:

```bash
# C1 — uma conexão nova por requisição
for i in 1 2 3 4 5 6 7 8 9 10; do
  curl --http1.1 -s -o /dev/null -H "Connection: close" http://<IP-servidor>:8080/notas.txt
done

# C2 — uma única conexão persistente para as 10 requisições
curl --http1.1 -s -o /dev/null \
  http://<IP-servidor>:8080/notas.txt http://<IP-servidor>:8080/notas.txt http://<IP-servidor>:8080/notas.txt \
  http://<IP-servidor>:8080/notas.txt http://<IP-servidor>:8080/notas.txt http://<IP-servidor>:8080/notas.txt \
  http://<IP-servidor>:8080/notas.txt http://<IP-servidor>:8080/notas.txt http://<IP-servidor>:8080/notas.txt \
  http://<IP-servidor>:8080/notas.txt
```

## Conteúdo de `www/`

Página de interoperabilidade (`index.html`) com CSS, JavaScript, PNG, JPEG, TXT, JSON e PDF.
