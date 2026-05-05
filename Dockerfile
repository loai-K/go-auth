FROM golang:1.26-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /bin/auth-server ./cmd/auth-server

FROM alpine:latest
RUN apk --no-cache add ca-certificates
RUN addgroup -S app && adduser -S -G app -u 10001 app
COPY --from=builder /bin/auth-server /usr/local/bin/auth-server
RUN chown app:app /usr/local/bin/auth-server && chmod 0755 /usr/local/bin/auth-server
USER app
WORKDIR /home/app
EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/auth-server"]
