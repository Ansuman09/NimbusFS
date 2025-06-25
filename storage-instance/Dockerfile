#Docker file for receivers all will use  the same image
#only difference will be the mapped ips
# Use the official Golang image as a base
FROM golang:1.24.3

# Set the working directory inside the container
WORKDIR /app

# Copy go.mod and go.sum files first (to leverage caching)
COPY go.mod ./

# Download dependencies
# RUN go mod download
RUN mkdir storage-instance
# Copy the entire project into the container
COPY . .

# Build the Go binary
# RUN go build -o receiver .

# Expose the desired port (adjust based on your app)
EXPOSE 9443

# Run the compiled binary
CMD ["go","run","receiver.go"]
