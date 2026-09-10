# Build stage: compile a static Go binary.
FROM golang:1.25-alpine AS builder
WORKDIR /app

# Cache module downloads separately from source changes.
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /server ./cmd/server

# Final stage: minimal runtime image, no Go toolchain.
FROM alpine:latest
# ca-certificates: needed for TLS to a managed Postgres (Supabase
# requires SSL) and any other outbound HTTPS calls.
RUN apk --no-cache add ca-certificates
COPY --from=builder /server /server

EXPOSE 8080
CMD ["/server"]
