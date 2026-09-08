test:
	@echo "Testing all modules"
	go test -v ./...

build:
	@echo "Building server and client"
	go build -o server ./cmd/server/*.go
	go build -o client ./cmd/client/*.go

run_db:
	docker compose up -d

down_db:
	docker compose down
