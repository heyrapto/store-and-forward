# Deployment Guide

This document outlines how to deploy the WhatsApp-style offline messaging service. It covers the local development setup, Docker deployment strategy, and the mechanics of simulating node failures.

## Local Development

### Prerequisites
- **Go 1.22+**
- **Node 20+**
- **Docker + Docker Compose**

### Running the System
The entire cluster (Gateways, Redis, Load Balancer) can be launched with a single command:
```bash
docker-compose up
```

### What Starts?
- **3 Gateway Nodes**: Running on internal ports or exposed optionally (e.g., `:8081`, `:8082`, `:8083`).
- **Load Balancer (Caddy)**: Exposed on `:80` and `:443` for client connections.
- **Redis**: Used for pub/sub (cross-gateway presence and routing).
- **SQLite Volume**: Shared persistent storage for the database.

### Environment Variables
Key configuration injected into the gateway services:
- `GATEWAY_ID`: Unique identifier for the gateway (e.g., `gw-1`).
- `REDIS_URL`: Connection string for the Redis instance.
- `SQLITE_PATH`: Path to the shared SQLite database file.
- `LISTEN_ADDR`: The HTTP/WS listen address for the gateway.
- `LOG_LEVEL`: Application logging level (`debug`, `info`, `warn`, `error`).

## Directory Layout for Deploy Files

The deployment configuration lives under the `server/deploy/` directory:
```
server/deploy/
├── docker-compose.yml  # Cluster definition
├── Caddyfile           # Reverse proxy & load balancing rules
└── Dockerfile          # Multi-stage build for the Gateway Go app
```

## Docker Compose Setup
The `docker-compose.yml` sets up a highly available test cluster:
- **Gateway**: Deployed as 3 replicas (or 3 named services `gateway-1`, `gateway-2`, `gateway-3`) connected to the internal network.
- **Redis**: An unmodified Redis container for pub/sub messaging.
- **Caddy**: The edge router mapping external traffic to internal gateways.
- **Volumes**: A `sqlite-data` volume is mounted to all gateway containers to share the SQLite DB.
- **Networks**: A custom bridge network links the services.

## Caddyfile
Caddy acts as our edge load balancer. It provides:
- **Round-Robin Reverse Proxy**: Distributes incoming HTTP/WS traffic evenly across the 3 gateway replicas.
- **WebSocket Support**: Automatically handles HTTP `Upgrade: websocket` headers.
- **TLS**: Auto-provisioned TLS (local certs via Caddy's internal CA for local dev).

## Dockerfile
The service uses an optimized multi-stage build:
1. **Builder Stage**: Uses a heavy Go image to download dependencies and compile the binary.
2. **Runtime Stage**: A minimal base image (like Alpine or distroless) that copies only the compiled binary.
3. **Security**: Runs as a non-root user for security best practices.

## Migrations
Database schema changes are managed by **Goose**. 
- On startup, the gateway entrypoint automatically runs `goose up` to apply migrations.
- **Migration 001** creates the initial tables and seeds the default `general` group, ensuring all users have a common room to chat in immediately.

## 'Kill a Gateway' Demo
To demonstrate the fault tolerance and state recovery of the system:
1. Start the cluster: `docker-compose up -d`.
2. Connect clients and verify connections in the **god-view** dashboard.
3. Stop one of the gateways mid-operation:
   ```bash
   docker-compose stop gateway2
   ```
4. **Observe**: Watch the connected devices drop and immediately reconnect to the surviving gateways (`gateway1` or `gateway3`).
5. **Verify**: Ensure no in-flight messages were lost. The mailbox sequence cursor ensures any un-acked messages are re-delivered upon reconnection.

## Production Notes
- **Shared SQLite Tradeoff**: For this demo, we use SQLite mounted via a shared Docker volume. In a true distributed production environment, SQLite on a shared filesystem is an anti-pattern (due to file locking and latency).
- **Scaling Up**: To scale beyond a single machine or avoid SQLite limitations, swap the repository interface implementation from SQLite to **PostgreSQL** or **Cassandra**. The application's domain logic is decoupled from the storage layer to allow this exact swap.
