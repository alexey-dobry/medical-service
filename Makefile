MODULES := internal/pkg internal/services/auth_service internal/services/user_service tests
SERVICES := auth_service user_service
PROTO_DIR := internal/pkg/proto
GEN_DIR := internal/pkg/gen

.PHONY: up down logs build vet test test-integration lint proto tidy

## up: build images and start the whole stack
up:
	docker compose up --build -d --wait

## down: stop the stack (add V=1 to drop volumes)
down:
	docker compose down $(if $(V),-v,)

## logs: follow logs of the Go services
logs:
	docker compose logs -f auth-service user-service

## build: go build in every module
build:
	@for m in $(MODULES); do echo "==> $$m"; (cd $$m && go build ./...) || exit 1; done

## vet: go vet in every module
vet:
	@for m in $(MODULES); do echo "==> $$m"; (cd $$m && go vet ./...) || exit 1; done

## test: unit tests in every module (no external dependencies)
test:
	@for m in $(MODULES); do echo "==> $$m"; (cd $$m && go test -race -count=1 ./...) || exit 1; done

## test-integration: full scenario against a running stack (make up first)
test-integration:
	cd tests && go test -tags=integration -count=1 -v ./...

lint:
	@for m in $(MODULES); do echo "==> $$m"; (cd $$m && golangci-lint run ./...) || exit 1; done

## proto: regenerate gRPC code (needs protoc, protoc-gen-go, protoc-gen-go-grpc)
proto:
	protoc -I $(PROTO_DIR) \
		--go_out=$(GEN_DIR) --go_opt=paths=source_relative \
		--go-grpc_out=$(GEN_DIR) --go-grpc_opt=paths=source_relative \
		$(PROTO_DIR)/auth/auth.proto $(PROTO_DIR)/user/user.proto

## tidy: go mod tidy in every module
tidy:
	@for m in $(MODULES); do echo "==> $$m"; (cd $$m && go mod tidy) || exit 1; done
