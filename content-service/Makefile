build:
	@go build -o ./bin/moderation-service ./main.go

test:
	go test -v ./...

run: build
	@./bin/identity-service
