FROM golang:1.25-alpine AS builder
WORKDIR /app
# Install dependencies
RUN apk add --no-cache git ca-certificates
# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download
# Copy source
COPY . .
# Build
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /currency ./currency/cmd/currency

# Runtime image
FROM alpine:3.19
RUN apk add --no-cache ca-certificates wget
WORKDIR /app
COPY --from=builder /currency .
EXPOSE 8081
CMD ["./currency"]
