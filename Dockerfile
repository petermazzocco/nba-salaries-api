FROM golang:1.23 AS builder
WORKDIR /app
# Copy go mod and sum files
COPY go.mod go.sum ./
# Download dependencies
RUN go mod download
# Copy the source code
COPY . .
# Build the application statically for Alpine
RUN mkdir -p bin && CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o bin/nba-salaries-api ./cmd/api

# Use a smaller image for the final container
FROM alpine:3.17
WORKDIR /app
# Install CA certificates for HTTPS requests
RUN apk --no-cache add ca-certificates
# Copy the binary from the builder stage
COPY --from=builder /app/bin/nba-salaries-api .
# Expose the port your API server listens on (adjust if needed)
EXPOSE 8080
# Run the API server
CMD ["./nba-salaries-api"]
