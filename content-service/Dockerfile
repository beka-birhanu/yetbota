FROM golang:1.26-alpine AS builder
WORKDIR /repo
COPY content-service/go.mod content-service/go.sum ./content-service/
COPY common/ ./common/
RUN cd content-service && go mod download
COPY content-service/ ./content-service/
RUN cd content-service && CGO_ENABLED=0 GOOS=linux go build -o /app/content-service ./cmd/main.go

FROM alpine:3.21
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /app
COPY --from=builder /app/content-service .
COPY content-service/config.yaml .
COPY content-service/migrations ./migrations
EXPOSE 9696 9966
CMD ["./content-service"]

