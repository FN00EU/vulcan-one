#Basic dockerfile
FROM golang:1.20.14 as builder

WORKDIR /app


COPY go.mod go.sum /app/
COPY cmd/vulcanone/main.go /app/cmd/vulcanone/main.go
COPY configs /app/configs
COPY internal /app/internal

RUN go mod download
RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux go build -o main ./cmd/vulcanone

FROM alpine:3.19.1


WORKDIR /app

COPY --from=builder /app/main /app/main
COPY --from=builder /app/configs /app/configs

EXPOSE 8080

CMD ["/app/main"]