build:
	@go build -o ./bin/content-service ./cmd/main.go

test:
	go test -v ./...

run: build
	@set -a && [ -f .env ] && . ./.env; set +a; ./bin/content-service
