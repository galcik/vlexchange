test:
	go test -v -cover ./...

generate:
	sqlc generate
	go generate -v ./...