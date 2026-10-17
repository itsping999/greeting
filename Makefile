SHELL := /bin/sh
TOOLS_DIR := $(CURDIR)/.tools/bin
PROTOC ?= protoc
DOCS_PYTHON ?= python3
GOVULNCHECK_VERSION := v1.9.0
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
GO_LDFLAGS := -X main.Version=$(VERSION)
API_PROTO_FILES := $(shell find api -name '*.proto' | sort)
export PATH := $(TOOLS_DIR):$(PATH)

.PHONY: init api generate all build run test smoke examples-smoke docker-build-linux-amd64 docker-build-linux-arm docker-build-linux-arm64
init:
	mkdir -p $(TOOLS_DIR)
	GOBIN=$(TOOLS_DIR) go install github.com/go-kratos/kratos/cmd/kratos/v2@v2.0.0-20260228034312-fe9258d38fd4
	GOBIN=$(TOOLS_DIR) go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.36.11
	GOBIN=$(TOOLS_DIR) go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.5.1
	GOBIN=$(TOOLS_DIR) go install github.com/go-kratos/kratos/cmd/protoc-gen-go-http/v2@v2.0.0-20260404020628-f149714c1d54
	GOBIN=$(TOOLS_DIR) go install github.com/google/gnostic/cmd/protoc-gen-openapi@v0.7.0
	GOBIN=$(TOOLS_DIR) go install github.com/google/wire/cmd/wire@v0.7.0
	GOBIN=$(TOOLS_DIR) go install github.com/envoyproxy/protoc-gen-validate@v1.1.0

api:
	$(PROTOC) --proto_path=./api --proto_path=./third_party --go_out=paths=source_relative:./api --go-http_out=paths=source_relative,omitempty=false:./api --go-grpc_out=paths=source_relative:./api --openapi_out=fq_schema_naming=true,default_response=false:. --validate_out=paths=source_relative,lang=go:./api $(API_PROTO_FILES)
config:
	@echo "Runtime configuration uses Go structs; no config generation required."
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

.PHONY: doctor task fmt fmt-check check-generated check-docs init-security vuln-check check
doctor:
	python3 scripts/dev.py doctor
task:
	$(DOCS_PYTHON) -B scripts/dev.py task --name "$(NAME)"
fmt:
	python3 scripts/dev.py fmt
fmt-check:
	python3 scripts/dev.py fmt-check
check-generated:
	python3 scripts/dev.py check-generated
check-docs:
	$(DOCS_PYTHON) -B docs/development/test_check_docs.py
	$(DOCS_PYTHON) -B scripts/test_dev.py
	$(DOCS_PYTHON) -B docs/development/check_docs.py
init-security:
	mkdir -p $(TOOLS_DIR)
	GOBIN=$(TOOLS_DIR) go install golang.org/x/vuln/cmd/govulncheck@$(GOVULNCHECK_VERSION)
vuln-check:
	$(TOOLS_DIR)/govulncheck -test ./...
check:
	$(MAKE) check-docs
	$(MAKE) fmt-check
	go vet ./...
	$(MAKE) test
	$(MAKE) check-generated
	$(MAKE) build

.PHONY: verify generated-check scaffold-proto scaffold-module e2e-smoke docker-build
verify: check
generated-check: check-generated
scaffold-proto:
	go run ./cmd/scaffold proto $(if $(filter 1,$(CRUD)),--crud,) "$(PROTO)"
scaffold-module:
	$(MAKE) scaffold-proto
	$(MAKE) generate
e2e-smoke: build
	./scripts/e2e-smoke.sh
docker-build:
	docker build -t greeting:local .

.PHONY: scaffold-smoke
scaffold-smoke:
	python3 scripts/dev.py scaffold-smoke
