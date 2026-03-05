build:
	@go build -o ./bin/identity-service ./main.go

test:
	go test -v ./...

run: build
	@./bin/identity-service
