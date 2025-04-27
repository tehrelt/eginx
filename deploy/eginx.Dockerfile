FROM golang:1.24.2-alpine AS builder

RUN apk update --no-cache
WORKDIR /app
COPY . /app
RUN go clean --modcache
RUN go build -o app ./cmd/app/main.go

FROM alpine

RUN apk update --no-cache
WORKDIR /app
COPY --from=builder /app /app

CMD ["./app"]