---
title: 🚀 Self-host
layout: default
permalink: /self-host
---

# Self-Hosting Your Project 🚀

This section explains how to deploy and run your project in a self-hosted environment. You have three main options:

1. **Using Docker directly with the provided Dockerfile 🐳**
2. **Using the Makefile for an automated build and run process ⚙️**
3. **Executing the Go code directly with command-line flags 👨‍💻**

---

## Option 1: Docker Deployment 🐳

The Docker deployment is based on the production Dockerfile located at `docker/Dockerfile.prod`. It produces a minimal, [distroless](https://github.com/GoogleContainerTools/distroless) image that runs the API as a non-root user.

### 1. **Prepare the Environment File 📄**

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

### 2. Build the Docker Image 🏗️

Run the following command in the root of your project to build the image:

```bash
docker build -f docker/Dockerfile.prod -t simpleauthlink .
```

### 3. Run the Docker Container 🚢

Once the image is built, start a container using the environment file:

```bash
docker run --name simpleauthlink --env-file .env -p ${PORT}:${PORT} simpleauthlink
```

The API listens on the `PORT` value defined in your `.env` (default `8080`), both inside the container and on your host.

---

## Option 2: Using the Makefile ⚙️

The Makefile in the root of the project simplifies the build and run process:

- `make api` — builds and runs the API container using `docker/Dockerfile.prod`.
- `make demo` — builds and runs the demo container using `docker/Dockerfile.demo`.
- `make swagger-ui` — generates the swagger spec and serves it locally with Swagger UI.
- `make clean-api` / `make clean-demo` — clean up the corresponding containers and images.

### 1. Set Up the Environment File 📋

As with the Docker approach, copy and edit `example.env` to `.env` with your configuration.

### 2. Build and Run Using Make 🏃‍♂️

Run the following command from your project root:

```bash
make api
```

This will:

- Clean up any previous containers and images.
- Build the Docker image using `docker/Dockerfile.prod`.
- Run the container with the environment variables from `.env` and expose the API on the configured `PORT`.

## Option 3: Running the Go Code Directly 👨‍💻

If you prefer to run the API without Docker, you can execute the Go code directly. The entry point is `cmd/authapi/main.go` and it accepts several flags, which can also be provided via environment variables:

- `--host` or environment variable `HOST` (default: `"0.0.0.0"`)
- `--port` or environment variable `PORT` (default: `8080`)
- `--secret` or environment variable `SECRET` (required)
- `--email-addr` or environment variable `EMAIL_ADDR` (required)
- `--email-user` or environment variable `EMAIL_USER` (SMTP username)
- `--email-pass` or environment variable `EMAIL_PASS` (SMTP password)
- `--email-host` or environment variable `EMAIL_HOST` (required)
- `--email-port` or environment variable `EMAIL_PORT` (default: `587`)
- `--queue-size` or environment variable `NOTIFICATION_QUEUE_SIZE` (default: `1000`)
- `--queue-workers` or environment variable `NOTIFICATION_QUEUE_WORKERS` (default: `10`)

The command loads `.env` automatically when it exists, so you can simply run:

```bash
go run ./cmd/authapi
```

Or pass everything explicitly:

```bash
go run ./cmd/authapi \
  --host="0.0.0.0" \
  --port=8080 \
  --secret="my_backend_secret" \
  --email-addr="test@test.com" \
  --email-user="test@test.com" \
  --email-pass="smtp_server_password" \
  --email-host="smtp.example.com" \
  --email-port=587
```

---

Choose the option that best fits your deployment needs. The Docker and Makefile options are ideal for containerized environments, while running the Go code directly might be preferred for development or custom setups.