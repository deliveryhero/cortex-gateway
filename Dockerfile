FROM --platform=$BUILDPLATFORM golang:1.26-alpine as builder
RUN apk update && apk add --no-cache git ca-certificates && update-ca-certificates

WORKDIR /app
COPY . .

ARG TARGETOS
ARG TARGETARCH

RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -o /go/bin/cortex-gateway

# executable image
FROM alpine:3
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /go/bin/cortex-gateway /go/bin/cortex-gateway

ENTRYPOINT ["/go/bin/cortex-gateway"]
