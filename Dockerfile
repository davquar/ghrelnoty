FROM golang:1.22 AS builder

RUN apt-get update && \
    apt-get install -y --no-install-recommends ca-certificates

ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

WORKDIR /app
COPY . .
RUN go build ./cmd/ghrelnoty

FROM scratch
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=builder /app/ghrelnoty /app/

WORKDIR /app
ENTRYPOINT ["./ghrelnoty"]