FROM golang:1.27.2-alpine AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
ARG TARGETOS=linux
ARG TARGETARCH
ARG VERSION=dev
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -ldflags "-s -w -X main.Version=${VERSION}" -o /out/server ./cmd/server

FROM scratch
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
WORKDIR /app
COPY --from=builder /out/server ./server
COPY configs/container.yaml ./config.yaml
USER 65532:65532
EXPOSE 8000 9000
ENTRYPOINT ["/app/server"]
CMD ["-conf", "/app/config.yaml"]
