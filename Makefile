test:
	go test --race ./...

lint:
	golangci-lint run ./...

run:
	go run main.go
