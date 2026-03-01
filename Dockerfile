FROM golang:1.24.4-alpine AS builder

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o burrow ./cmd/burrow

FROM alpine:3.21.3

RUN apk add --no-cache ca-certificates

COPY --from=builder /build/burrow /usr/local/bin/burrow

ENTRYPOINT ["burrow"]
