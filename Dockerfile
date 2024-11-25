# Stage 1: Build the Go application
FROM golang:latest AS build

# Set the working directory inside the container
WORKDIR /app

# Copy the Go module files and the source code
COPY go.mod go.sum agent.go ./

# Download dependencies
RUN go mod download

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o agent agent.go

FROM scratch

# Copy the statically compiled Go binary from the build stage
COPY --from=build /app/agent /app/

# Copy CA certificates from the build stage (from Golang base image)
COPY --from=build /etc/ssl/certs /etc/ssl/certs

# Expose the necessary port
EXPOSE 8080

# Set the command to run the Go application
CMD ["/app/agent", "-interval=15"]
