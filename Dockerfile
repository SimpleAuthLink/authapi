# build
FROM golang:1.21-alpine as builder

WORKDIR /app/data
COPY . .

RUN go mod tidy
RUN go build -o /authapi ./cmd/authapi/main.go

# deploy
FROM alpine:latest

WORKDIR /
COPY --from=builder /authapi /authapi
COPY --from=builder /app/data/assets /assets

ENTRYPOINT /authapi