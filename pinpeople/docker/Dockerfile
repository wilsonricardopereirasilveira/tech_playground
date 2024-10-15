# Build stage
FROM golang:1.23-alpine AS builder

# Install Git
RUN apk add --no-cache git

WORKDIR /app

# Copy go.mod and go.sum (if they exist)
COPY go.mod ./
COPY go.sum* ./

# Download dependencies
RUN go mod download

# Copy the rest of the source code
COPY . .

# Copy the wait-for-db.sh script
COPY wait-for-db.sh ./wait-for-db.sh

# Set execution permissions for the wait-for-db.sh script
RUN chmod +x ./wait-for-db.sh

# Build the main application
RUN go build -o main ./cmd/api

# Build the importer
RUN go build -o import ./cmd/import

# Final stage
FROM alpine:latest

# Install PostgreSQL client and bash
RUN apk add --no-cache postgresql-client bash

WORKDIR /root/

# Copy the compiled executables
COPY --from=builder /app/main .
COPY --from=builder /app/import .

# Copy the wait-for-db script
COPY --from=builder /app/wait-for-db.sh .

# Copy the CSV file
COPY data/employees.csv ./data/

# Make the wait-for-db.sh script executable
RUN chmod +x /root/wait-for-db.sh

# Expose port 8080
EXPOSE 8080

# Default command to run the main application
CMD ["./main"]