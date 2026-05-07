FROM golang:1.21-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o nvrhikvision-exporter ./cmd/exporter

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy binary from builder
COPY --from=builder /app/nvrhikvision-exporter .

# Copy config example
COPY config.yaml .

EXPOSE 9102

ENTRYPOINT ["./nvrhikvision-exporter"]
CMD ["-config=config.yaml"]
