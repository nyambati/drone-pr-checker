# Start from the latest golang base image
FROM golang:latest AS builder

# Add Maintainer Info
LABEL maintainer="12892110+nyambati@users.noreply.github.com"

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./


# Copy the source from the current directory to the working Directory inside the container
COPY . .

# Build the Go app
RUN go mod download && CGO_ENABLED=0 go build -o drone-pr-checker .

# Start a new stage from scratch
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /plugin/

# Copy the Pre-built binary file from the previous stage
COPY --from=builder /app/drone-pr-checker .

# Expose port 8080 to the outside world
EXPOSE 8080
.r
# Command to run the executable
CMD ["./drone-pr-checker"]
