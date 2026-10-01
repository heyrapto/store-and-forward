# System Architecture

## 1. System Overview

The system is designed to simulate a WhatsApp-style offline messaging platform. It consists of multiple client "React phones" and a "god-view" admin interface, routing traffic through Caddy as a reverse proxy/load balancer, into Go-based gateway servers. The state is durably persisted in a shared SQLite database (WAL mode), while transient state (presence and cross-gateway routing) is managed by Redis. 

```mermaid
flowchart TD
    subgraph Clients
        R1[React Phone 1]
        R2[React Phone 2]
        GV[God View Admin]
    end

    Caddy[Caddy Reverse Proxy]

    subgraph Gateways
        GW1[Go Gateway 1]
        GW2[Go Gateway 2]
    end

    subgraph Storage
        SQLite[(SQLite DB)]
        Redis[(Redis)]
    end

    R1 -->|WSS| Caddy
    R2 -->|WSS| Caddy
    GV -->|HTTPS/SSE| Caddy

    Caddy --> GW1
    Caddy --> GW2

    GW1 <--> SQLite
    GW1 <--> Redis
    GW2 <--> SQLite
    GW2 <--> Redis
    
    GW1 <..>|Pub/Sub| GW2
```

## 2. Clean Architecture

The Go backend strictly adheres to Clean Architecture principles. The central rule is the **Dependency Rule**: source code dependencies must point only inward, toward higher-level policies. 

- **Domain**: Contains enterprise-wide business rules, entities, and state machines.
- **Use Cases / Ports**: Contains application-specific business rules and defines interfaces (ports) for external agencies.
- **Adapters**: Converts data from the format most convenient for the use cases and entities, to the format most convenient for external agencies (e.g., DB, Web).
- **Main**: The entry point that wires all dependencies together (dependency injection).

## 3. Directory Structure

```text
server/
├── cmd/
│   └── gateway/
│       └── main.go               # Wiring, Config, Entrypoint
├── internal/
│   ├── domain/                   # Entities, Events, State Machines
│   │   ├── device.go
│   │   ├── message.go
│   │   ├── delivery.go
│   │   └── group.go
│   ├── usecase/                  # Application Business Logic
│   │   ├── ports/                # Interfaces (Repos, Stores, Bus)
│   │   │   ├── repository.go
│   │   │   └── presence.go
│   │   ├── connect_device.go
│   │   ├── send_message.go
│   │   └── ack_delivery.go
│   ├── adapter/                  # Implementations of Ports
│   │   ├── http/                 # REST & SSE
│   │   ├── websocket/            # WS Transport
│   │   ├── sqlite/               # Persistence
│   │   ├── redis/                # Presence & PubSub Bus
│   │   └── memory/               # Connection Registry
│   └── config/                   # Environment loading
```

## 4. Domain Layer

The Domain layer is purely independent. 
- **Device**: Represents a physical hardware client.
- **Message**: The core payload entity.
- **Group**: Represents a collection of devices.
- **Delivery**: Represents the edge between a Message and a Recipient Device.
- **Delivery State Machine**: Tracks message status progression: `Pending` → `Delivered` → `Read`.
- **Domain Events**: Event payloads (e.g., `MessageSentEvent`, `DeliveryAckEvent`) used to trigger side-effects without tight coupling.

## 5. Use Cases & Ports

### Use Cases
- `register_device`: Issues credentials for a new device.
- `connect_device`: Handles the WebSocket handshake and synchronizes offline messages.
- `disconnect_device`: Cleans up local registry and presence.
- `send_message`: Processes inbound messages and initiates fan-out.
- `ack_delivery`: Transitions delivery state to 'Delivered'.
- `mark_read`: Transitions delivery state to 'Read'.
- `sweep_expired`: Background job to clear messages older than 30 days.

### Ports
- `DeviceRepo`: CRUD for devices.
- `MessageRepo`: CRUD for messages.
- `DeliveryRepo`: CRUD and state updates for deliveries.
- `PresenceStore`: Tracks online/offline status (Redis/Memory).
- `EventBus`: Publishes domain events.
- `ConnectionRegistry`: Tracks active local TCP connections.
- `Clock`: Time interface for testability.
- `IDGenerator`: ULID generator interface.

## 6. Adapters

- **Transport**:
  - `WebSocket`: Main bidirectional stream for devices (`coder/websocket`).
  - `HTTP REST`: Device registration endpoints.
  - `Admin SSE`: Server-Sent Events for the God-View dashboard.
- **Persistence**: `modernc.org/sqlite` (pure Go, CGO-free, WAL mode for concurrency).
- **Presence & Bus**: Inter-node transient state via Redis (or in-memory for single node).
- **Registry**: Thread-safe, sharded connection registry for O(1) local connection lookups.

## 7. Connection Internals

Active connections map to a pair of goroutines:
1. **Reader Goroutine**: Blocking reads from the WS connection. Unmarshals JSON, invokes use cases.
2. **Writer Goroutine**: Selects from a bounded outbound channel. Enforces backpressure.
- **Sharded Registry**: 64 shards using a hash of the DeviceID to minimize lock contention when storing pointers to outbound channels.
- **Heartbeat & Presence**: The client sends periodic PINGs. The reader goroutine updates a Redis presence key with a TTL (e.g., 30s) to indicate global online status.

## 8. Fan-out

When a group message is sent:
1. The use case identifies all group members.
2. It writes one row to the `messages` table and $N$ rows to the `deliveries` table within a single transaction.
3. A Worker Pool is notified to process the deliveries.
4. For each delivery, it checks presence. If online locally, it pushes to the local writer channel. If online remotely, it publishes to the Event Bus (Redis).

## 9. Cross-Gateway Routing (Episode 3 Preview)

To support multiple gateway nodes:
1. Gateway A receives a message for Device X.
2. Device X is connected to Gateway B.
3. Gateway A checks the Redis `PresenceStore` and discovers Device X is on node `GW-B`.
4. Gateway A publishes the delivery payload to Redis Pub/Sub topic `gateway:GW-B`.
5. Gateway B consumes the topic and pushes the payload directly into Device X's bounded outbound channel.
