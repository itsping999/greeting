SHELL := /bin/sh
TOOLS_DIR := $(CURDIR)/.tools/bin
PROTOC ?= protoc
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
GO_LDFLAGS := -X main.Version=$(VERSION)
API_PROTO_FILES := $(shell find api -name '*.proto' | sort)
INTERNAL_PROTO_FILES := $(shell find internal -name '*.proto' | sort)
export PATH := $(TOOLS_DIR):$(PATH)

.PHONY: init api config generate all build run test smoke examples-smoke docker-build-linux-amd64 docker-build-linux-arm docker-build-linux-arm64
init:
	mkdir -p $(TOOLS_DIR)
	GOBIN=$(TOOLS_DIR) go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.36.11
	GOBIN=$(TOOLS_DIR) go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.5.1
	GOBIN=$(TOOLS_DIR) go install github.com/go-kratos/kratos/cmd/protoc-gen-go-http/v2@v2.0.0-20260404020628-f149714c1d54
	GOBIN=$(TOOLS_DIR) go install github.com/google/gnostic/cmd/protoc-gen-openapi@v0.7.0
	GOBIN=$(TOOLS_DIR) go install github.com/google/wire/cmd/wire@v0.7.0
	GOBIN=$(TOOLS_DIR) go install github.com/envoyproxy/protoc-gen-validate@v1.1.0

api:
	$(PROTOC) --proto_path=./api --proto_path=./third_party --go_out=paths=source_relative:./api --go-http_out=paths=source_relative:./api --go-grpc_out=paths=source_relative:./api --openapi_out=fq_schema_naming=true,default_response=false:. --validate_out=paths=source_relative,lang=go:./api $(API_PROTO_FILES)
config:
	$(PROTOC) --proto_path=./internal --proto_path=./third_party --go_out=paths=source_relative:./internal $(INTERNAL_PROTO_FILES)
generate:
	cd cmd/server && wire
all:
	$(MAKE) api
	$(MAKE) config
	$(MAKE) generate
build:
	mkdir -p bin
	go build -ldflags "$(GO_LDFLAGS)" -o bin/server ./cmd/server
	go build -o bin/client ./cmd/client
run:
	go run ./cmd/server -conf ./configs/config.yaml
test:
	go test -race ./...
smoke: build
	./scripts/smoke.sh
examples-smoke: build
	./scripts/smoke.sh --brokers

docker-build-linux-amd64:
	mkdir -p bin/amd64
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "$(GO_LDFLAGS)" -o bin/amd64/app ./cmd/server
docker-build-linux-arm:
	mkdir -p bin/arm
	CGO_ENABLED=0 GOOS=linux GOARCH=arm GOARM=7 go build -ldflags "$(GO_LDFLAGS)" -o bin/arm/app ./cmd/server
docker-build-linux-arm64:
	mkdir -p bin/arm64
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -ldflags "$(GO_LDFLAGS)" -o bin/arm64/app ./cmd/server

.PHONY: doctor task fmt fmt-check check-generated check
doctor:
	python3 scripts/dev.py doctor
task:
	python3 scripts/dev.py task --name "$(NAME)"
fmt:
	python3 scripts/dev.py fmt
fmt-check:
	python3 scripts/dev.py fmt-check
check-generated:
	python3 scripts/dev.py check-generated
check:
	$(MAKE) fmt-check
	go vet ./...
	$(MAKE) test
	$(MAKE) check-generated
	$(MAKE) build
