test:
	go test -v -cover ./...

generate:
	sqlc generate
	go generate -v ./...

server:
	go run cmd/server/main.go