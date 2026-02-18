# Build stage
FROM golang:1.24-alpine AS builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -o vreco .

# Run stage
FROM alpine:3.21
RUN apk add --no-cache ca-certificates
WORKDIR /app

COPY --from=builder /app/vreco .
COPY templates ./templates
COPY static ./static
COPY posts ./posts

EXPOSE 8080
CMD ["./vreco"]
