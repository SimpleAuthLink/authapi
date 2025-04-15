[![Last release](https://img.shields.io/github/v/release/simpleauthlink/authapi?color=purple)](https://github.com/simpleauthlink/authapi/releases/latest)
[![GoDoc](https://godoc.org/github.com/simpleauthlink/authapi?status.svg)](https://godoc.org/github.com/simpleauthlink/authapi) 
[![Go Report Card](https://goreportcard.com/badge/github.com/simpleauthlink/authapi)](https://goreportcard.com/report/github.com/simpleauthlink/authapi)
[![Build and Test](https://github.com/simpleauthlink/authapi/actions/workflows/main.yml/badge.svg?branch=main)](https://github.com/simpleauthlink/authapi/actions/workflows/main.yml)
[![license](https://img.shields.io/github/license/simpleauthlink/authapi)](LICENSE)


<img src="https://simpleauth.link/assets/logo.svg" width=100/>

# SimpleAuth.link API

> Passwordless authentication for your users using just an email address.

This repository contains the source code of the SimpleAuth.link API Service.

Read full [documentation here](https://docs.simpleauth.link).

---

## Technical Details 💻

### Token Generation Process 🔑 

* By leveraging the Ed25519 signature algorithm, the service deterministically generates a private key using your App ID and secret. 
* This ensures that each token is cryptographically secure and uniquely tied to your application, eliminating the need to store sensitive keys.

### Stateless Architecture 🕊️

* SimpleAuth.link works without a traditional database. It does not store any user data, including email addresses, on its servers. 
* Instead, the data generated is self-contained, requiring no further information or state to be used. This stateless design increases security and reduces the risk of data breaches.

## Development 🧑‍💻

### Prerequisites 📝
 - Go (version 1.24 or later recommended)
 - Docker (optional, for containerized deployment)

### Clone the Repository 📥
```sh
git clone --branch v2 https://github.com/SimpleAuthLink/authapi.git
cd authapi
```

### Code Structure 🪜
* Modularity: The repository is structured into packages to simplify testing and future expansion.
  - `api/`: Contains the core API endpoint definitions and routing logic.
  - `cmd/`: Entry points and command-line utilities for running the API server.
  - `docker/`: Dockerfiles and configuration for containerized builds.
  - `internal/`: Internal packages and libraries used across the application.
  - `notification/`: Code handling user notifications.
  - `token/`: Modules for token creation and management.
  - `.github/`: GitHub-specific workflows and configurations (CI/CD, issue templates, etc.).
  
  This layout supports modular development and clear separation of concerns across different parts of the API service.

* Testing: Use the standard Go testing framework to run tests. You can run tests with:
   ```sh
   go test ./...
   ```

### Run with go 🦫
For development purposes, you can run the API server directly with Go. 
```sh
go run ./cmd/authapi -h
```
```sh
Usage of authapi:
  -email-addr string
        email account address
  -email-host string
        email server host
  -email-pass string
        email account password
  -email-port int
        email server port (default 587)
  -host string
        service host (default "0.0.0.0")
  -port int
        service host (default 8080)
  -secret string
        secret used to generate the tokens (default "simpleauthlink-secret")
```

### Run with docker 🐳

1. **Prepare the Environment File**
   
   Copy the `example.env` file to `.env` and edit the file to fill in your parameters:
   ```bash
   HOST="localhost"
   PORT=8080
   EMAIL_ADDR="test@test.com"
   EMAIL_PASS="smtp_server_password"
   EMAIL_HOST="smtp.example.com"
   EMAIL_PORT=587
   SECRET="my_backend_secret"
   ```

2. **Build the Docker Image**
   
   Run the following command in the root of your project to build the image:
   ```bash
   docker build -f docker/Dockerfile.prod -t simpleauthlink .
   ```

3. **Run the Docker Container**
   
   Once the image is built, start a container using the environment file:
   ```bash
   docker run --name simpleauthlink --env-file .env -p 8080:80 simpleauthlink
   ```
   This command maps the container’s port 80 to port 8080 on your host.