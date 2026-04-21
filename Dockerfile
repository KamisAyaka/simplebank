FROM golang:1.26-alpine3.23 AS builder
WORKDIR /app
COPY . .
RUN go build -o main main.go

FROM alpine:3.23
WORKDIR /app
ENV GIN_MODE=release
COPY --from=builder /app/main .
COPY app.docker.env ./app.env
COPY start.sh .
COPY db/migration ./db/migration

EXPOSE 8080
CMD ["/app/main"]
ENTRYPOINT [ "/app/start.sh" ]
