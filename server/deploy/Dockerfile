FROM golang:1.22-alpine AS builder

WORKDIR /app
# Install gcc/musl-dev if we were using cgo, but we use pure-go modernc.org/sqlite
# Wait, for mattn/go-sqlite3 (which we ended up using for Episode 1), CGO IS REQUIRED.
RUN apk add --no-cache gcc musl-dev

COPY go.mod go.sum ./
RUN go mod download

COPY . .
# Build with CGO enabled for mattn/go-sqlite3
RUN CGO_ENABLED=1 GOOS=linux go build -a -o gateway ./cmd/gateway

FROM alpine:3.19
WORKDIR /app
RUN apk add --no-cache tzdata
COPY --from=builder /app/gateway .
COPY --from=builder /app/migrations ./migrations

# SQLite database will be stored in /data
VOLUME /data
ENV DB_PATH=/data/messaging.db
ENV PORT=8080

EXPOSE 8080
ENTRYPOINT ["/app/gateway"]
