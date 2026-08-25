[![Last release](https://img.shields.io/github/v/release/simpleauthlink/authapi?color=purple)](https://github.com/simpleauthlink/authapi/releases/latest)
[![GoDoc](https://godoc.org/github.com/simpleauthlink/authapi?status.svg)](https://godoc.org/github.com/simpleauthlink/authapi)
[![Go Report Card](https://goreportcard.com/badge/github.com/simpleauthlink/authapi)](https://goreportcard.com/report/github.com/simpleauthlink/authapi)
[![Build and Test](https://github.com/simpleauthlink/authapi/actions/workflows/main.yml/badge.svg?branch=main)](https://github.com/simpleauthlink/authapi/actions/workflows/main.yml)
[![license](https://img.shields.io/github/license/simpleauthlink/authapi)](LICENSE)

<img src="https://simpleauth.link/assets/logo.svg" width=100/>

# SimpleAuth.link API

> Passwordless authentication for your users using just an email address.

This repository contains the source code of the SimpleAuth.link API Service.

- Read the full [documentation here](https://docs.simpleauth.link).
- Try the API in the interactive [Swagger UI](https://simpleauthlink.github.io/authapi/api/).

---

## Technical Details 💻

### Token Generation Process 🔑

- By leveraging the Ed25519 signature algorithm, the service deterministically generates a private key using your App ID and secret.
- This ensures that each token is cryptographically secure and uniquely tied to your application, eliminating the need to store sensitive keys.

### Stateless Architecture 🕊️

- SimpleAuth.link works without a traditional database. It does not store any user data, including email addresses, on its servers.
- Instead, the data generated is self-contained, requiring no further information or state to be used. This stateless design increases security and reduces the risk of data breaches.

### Token Structure 🧩

A token is composed of three parts — the expiration time, the signature and the user's email — base64url-encoded and separated by dots. Read more in the [documentation](https://docs.simpleauth.link/about/tokens).

## Development 🧑‍💻

### Prerequisites 📝

- Go (the version declared in [go.mod](./go.mod))
- Docker (optional, for containerized deployment)

### Clone the Repository 📥

```sh
git clone https://github.com/SimpleAuthLink/authapi.git
cd authapi
```

### Code Structure 🪜

- `api/`: Contains the core API endpoint definitions and routing logic, and the Go client in `api/client`.
- `cmd/`: Entry points and command-line utilities for running the API server (`authapi`) and the demo (`demo`).
- `docker/`: Dockerfiles for containerized builds.
- `notification/`: Code handling user notifications.
- `token/`: Modules for token creation and management.
- `scripts/`: Helper scripts (e.g. swagger generation).
- `.github/`: GitHub-specific workflows and configurations (CI/CD, etc.).

### Testing

```sh
go test ./...
```

### Run with go 🦫

For development purposes, you can run the API server directly with Go. The entry point is `cmd/authapi/main.go` and it loads `.env` automatically when present:

```sh
go run ./cmd/authapi
```

The main flags (also configurable via environment variables):

- `--host` (env `HOST`) — service host (default `0.0.0.0`)
- `--port` (env `PORT`) — service port (default `8080`)
- `--secret` (env `SECRET`) — secret used to generate the tokens
- `--email-addr` (env `EMAIL_ADDR`) — sender account address
- `--email-user` (env `EMAIL_USER`) — SMTP username
- `--email-pass` (env `EMAIL_PASS`) — SMTP password
- `--email-host` (env `EMAIL_HOST`) — SMTP server host
- `--email-port` (env `EMAIL_PORT`) — SMTP server port (default `587`)
- `--queue-size` (env `NOTIFICATION_QUEUE_SIZE`) — notification queue size (default `1000`)
- `--queue-workers` (env `NOTIFICATION_QUEUE_WORKERS`) — notification queue workers (default `10`)

### Run with docker 🐳

1. **Prepare the Environment File**

   Copy the `example.env` file to `.env` and edit the file to fill in your parameters:

   ```bash
   HOST="localhost"
   PORT=8080
   EMAIL_ADDR="test@test.com"
   EMAIL_USER="test@test.com"
   EMAIL_PASS="smtp_server_password"
   EMAIL_HOST="smtp.example.com"
   EMAIL_PORT=587
   SECRET="my_backend_secret"
   NOTIFICATION_QUEUE_SIZE=1000
   NOTIFICATION_QUEUE_WORKERS=10
   ```

2. **Build the Docker Image**

   ```bash
   docker build -f docker/Dockerfile.prod -t simpleauthlink .
   ```

3. **Run the Docker Container**

   ```bash
   docker run --name simpleauthlink --env-file .env -p ${PORT}:${PORT} simpleauthlink
   ```

   The API listens on the `PORT` value from your `.env` (default `8080`), both inside the container and on your host.

### Go client

The `api/client` package provides a typed Go client for the API:

```go
package main

import (
	"log"

	"github.com/simpleauthlink/authapi/api/client"
	"github.com/simpleauthlink/authapi/token"
)

func main() {
	cli, err := client.Default("<your-app-id>", "<your-app-secret>", false)
	if err != nil {
		log.Fatal(err)
	}

	// request a login token for a user
	if err := cli.RequestToken(new(token.Email).SetString("user@example.com")); err != nil {
		log.Fatal(err)
	}
}
```