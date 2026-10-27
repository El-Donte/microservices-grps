.PHONY: version
version:
	go version
	protoc --version

.PHONY: generate
generate:
	protoc -I ./ \
	--go_out ./pkg --go_opt paths=source_relative \
	--go-grpc_out ./pkg --go-grpc_opt paths=source_relative \
	api/product-service/product.proto

.PHONY: run-all
run-all:
	go run ./cmd/product-service & \
    go run ./cmd/cart

.PHONY: run-cart
run-cart:
	go run ./cmd/cart

.PHONY: run-product
run-product:
	go run ./cmd/product-service

.PHONY: test
test:
	go test ./...

.PHONY: test-race
test-race:
	go test -race ./...

.PHONY: test-api
test-api:
	go test -v -count=1 -tags=api ./test/