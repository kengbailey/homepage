
# Set the base image to Alpine Linux
FROM golang:1.22.4-bullseye

# Install Go runtime and set environment variables
RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates && \
    update-ca-certificates
# ENV GOROOT=/usr/lib/go \
#     GOPATH=/gopath \
#     GOBIN=/gopath/bin \
#     PATH=$PATH:$GOROOT/bin:$GOPATH/bin

# Copy the current directory contents into the container at /app
COPY . /app
WORKDIR /app

# Build and install the Go application
RUN go build -o main .

# Expose port 80 to the outside world
EXPOSE 1111

# Run the binary when the container launches
CMD ["/app/main"]
