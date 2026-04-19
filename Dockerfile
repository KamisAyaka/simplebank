FROM golang:1.26-alpine3.23 AS builder
WORKDIR /app
COPY . .
RUN go build -o main main.go
RUN apk add curl
RUN  curl -L https://github.com/golang-migrate/migrate/releases/download/v4.19.1/migrate.linux-amd64.tar.gz | tar xz

FROM alpine:3.23
WORKDIR /app
ENV GIN_MODE=release
COPY --from=builder /app/main .
COPY --from=builder /app/migrate ./migrate
COPY app.docker.env ./app.env
COPY start.sh .
COPY db/migration ./migration

EXPOSE 8080
CMD ["/app/main"]
ENTRYPOINT [ "/app/start.sh" ]
