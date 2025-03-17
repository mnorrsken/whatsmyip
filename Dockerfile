# Use golang:latest as the build image
FROM golang:latest AS builder

# Set the working directory to /app
WORKDIR /app

# Copy the Go module files and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the application code
COPY . .

# Build the Go application
RUN go build -o /whatsmyip

# Use scratch as the parent image
FROM scratch

# Copy the built application from the build image to the parent image
COPY --from=builder /whatsmyip /whatsmyip

# Set the entry point to run the application
ENTRYPOINT ["/whatsmyip"]
