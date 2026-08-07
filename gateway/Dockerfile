FROM golang:1.26-alpine AS builder
WORKDIR /src
RUN apk add --no-cache ca-certificates
COPY go.mod go.sum ./
COPY vendor ./vendor
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -mod=vendor -o /bin/service ./cmd

FROM alpine:3.21
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY --from=builder /bin/service /app/service
EXPOSE 8080
ENTRYPOINT ["/app/service"]
