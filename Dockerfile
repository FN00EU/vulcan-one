#Basic dockerfile
FROM golang:1.20.7 as builder

WORKDIR /app

COPY cmd/vulcanone/main.go internal go.mod go.sum configs /app/

RUN go mod download \
    && go mod tidy \
    && go build -o main ./cmd/vulcanone

FROM alpine:3.19.1

WORKDIR /app

COPY --from=builder /app/main /app/main

EXPOSE 8080

CMD ["/app/main"]
