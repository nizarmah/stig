# Create builder image.
FROM golang:1.24.3-alpine as builder

# Setup working directory
WORKDIR /src
COPY . .

# Install dependencies.
RUN go mod download && go mod verify

# Build the binary.
RUN go build -o game-record ./cmd/record

# Create a runner image.
FROM alpine:latest as runner

# Setup working directory.
WORKDIR /app
COPY --from=builder /src/game-record .

# Run the binary.
ENTRYPOINT ["./game-record"]
