.PHONY: compile
compile: ## Compile the proto file.
	protoc -I pkg/api/proto/ pkg/api/proto/geofences.proto --go_out=plugins=grpc:pkg/api/proto/

.PHONY: run
run: ## Build and run server.
#   	go build -race -ldflags "-s -w" -o bin/server cmd/server/main.go
#    	bin/server