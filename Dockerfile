# Stage 1: Build the application
FROM golang:1.25rc1-bookworm AS builder

# Set the working directory
WORKDIR /app

# Copy the application code into the container
COPY . .

# Build the application
RUN go build -o chunk

# Stage 2: Run the application
FROM debian:bookworm

# Set the working directory
WORKDIR /app

# Copy the compiled binary from the builder stage
COPY --from=builder /app/chunk .

# Run the application in a continuous loop
CMD while true; do sleep 30; done
