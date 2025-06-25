# Use the official Go base image
FROM golang:1.23

# Set the working directory inside the container
WORKDIR /app

# Copy go.mod and go.sum
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the rest of your source code
COPY . .

# Build your app (replace `main.go` with your actual entry file if needed)
RUN go build -o server .

# Expose the port Render will provide via PORT env variable
EXPOSE 8080

# Run the server (PORT will be set by Render)
CMD ["./server"]
