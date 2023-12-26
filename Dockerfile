# Use the official Go image as the base image
FROM golang:1.21.5-alpine

# Set the working directory inside the container
WORKDIR /app

# Copy the source code into the container
COPY . .

# Build the Go application
RUN go build -o app

EXPOSE 9085
# Set the entry point for the container
ENTRYPOINT ["./app"]
