# Use golang:latest as the build image
FROM --platform=$BUILDPLATFORM golang:1.24 AS builder

# Set ARGs for multi-arch builds
ARG TARGETOS
ARG TARGETARCH

# Set the working directory to /app
WORKDIR /app

# Copy the Go module files and download dependencies
# COPY go.mod go.sum ./
# RUN go mod download

# Copy the rest of the application code
COPY . .

# Build the Go application with multi-arch support
RUN GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -o /whatsmyip

# Use scratch as the parent image
FROM scratch

# Copy the built application from the build image to the parent image
COPY --from=builder /whatsmyip /whatsmyip
COPY --from=builder /templates /templates

# Set the command to run the application (changed from ENTRYPOINT)
CMD ["/whatsmyip"]
