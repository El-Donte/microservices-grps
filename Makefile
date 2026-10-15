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