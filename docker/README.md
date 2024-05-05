## Start services
Execute this command from this directory to start the example services:

`docker compose up -d`

## Call nginx

Open from browser:

http://localhost:8000/example

Or use command line curl:

`curl http://127.0.0.1:8000/example`

## Watch logs

Execute this command from this directory:

`docker compose logs -f traefik`

### You will see something like this:

```
traefik-1  | [nginx] 2024/05/05 19:52:41 172.21.0.1:56300 GET /example: 404 Not Found HTTP/1.1
traefik-1  | 
traefik-1  | Request Headers:
traefik-1  | X-Forwarded-Proto: http
traefik-1  | X-Forwarded-Port: 8000
traefik-1  | X-Forwarded-Host: 127.0.0.1:8000
traefik-1  | X-Forwarded-Server: 2f410295c1c0
traefik-1  | X-Real-Ip: 172.21.0.1
traefik-1  | User-Agent: curl/7.81.0
traefik-1  | Accept: */*
traefik-1  | 
traefik-1  | Response Headers:
traefik-1  | Content-Type: text/html
traefik-1  | Server: nginx/1.25.5
traefik-1  | Date: Sun, 05 May 2024 19:52:41 GMT
traefik-1  | Content-Length: 153
traefik-1  | 
traefik-1  | Response Content Length: 153
traefik-1  | 
traefik-1  | Response Body:
traefik-1  | <html>
traefik-1  | <head><title>404 Not Found</title></head>
traefik-1  | <body>
traefik-1  | <center><h1>404 Not Found</h1></center>
traefik-1  | <hr><center>nginx/1.25.5</center>
traefik-1  | </body>
traefik-1  | </html>
traefik-1  | 
traefik-1  | 
```