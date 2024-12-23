FROM golang:1.23.4 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o main .

# Minimal base image
FROM alpine:latest

WORKDIR /app

# Copies binary from the builder stage
COPY --from=builder /app/main .

EXPOSE 8080

CMD ["./main"]

