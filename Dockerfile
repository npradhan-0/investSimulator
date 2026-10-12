
FROM golang:1.27-alpine AS builder

WORKDIR /app

# Copy dependency files 
COPY go.mod go.sum ./
RUN go mod download

# Copy the Go source code
COPY main.go ./

# Compile the Go application
RUN go build -o investSimulator main.go

# Use a small image to run the application
FROM alpine:latest

WORKDIR /app

# Copy the compiled application
COPY --from=builder /app/investSimulator .

# Run the application
CMD ["./investSimulator"]
