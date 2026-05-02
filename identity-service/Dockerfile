FROM golang:1.26-alpine AS builder
WORKDIR /repo
COPY identity-service/go.mod identity-service/go.sum ./identity-service/
COPY common/ ./common/
RUN cd identity-service && go mod download
COPY identity-service/ ./identity-service/
RUN cd identity-service && CGO_ENABLED=0 GOOS=linux go build -o /app/identity-service ./cmd/main.go

FROM alpine:3.21
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /app
COPY --from=builder /app/identity-service .
COPY identity-service/config.yaml .
COPY identity-service/migrations ./migrations
EXPOSE 6969 6699
CMD ["./identity-service"]

