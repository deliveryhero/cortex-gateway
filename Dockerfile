FROM golang:1.19-alpine as builder
RUN apk update && apk add --no-cache git ca-certificates && update-ca-certificates

WORKDIR /app
COPY . .

RUN CGO_ENABLED=0 go build -o /go/bin/mimir-gateway

# executable image
FROM alpine:3
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /go/bin/mimir-gateway /go/bin/mimir-gateway

ENTRYPOINT ["/go/bin/mimir-gateway"]
