# API Reference

This document covers both the REST API for management and dashboard interfaces, and the WebSocket API for real-time messaging.

## REST API

### `POST /devices`
Register a new device on the network.

- **Description:** Creates a new device identity and returns the assigned ID.
- **Request Body:**
  ```json
  {
    "name": "Bob's iPhone",
    "avatar": "https://example.com/avatar.png"
  }
  ```
- **Response Body:**
  ```json
  {
    "id": "dev-12345",
    "name": "Bob's iPhone",
    "avatar": "https://example.com/avatar.png",
    "created_at": "2026-10-01T10:00:00Z"
  }
  ```
- **Example cURL:**
  ```bash
  curl -X POST http://localhost:8080/devices \
    -H "Content-Type: application/json" \
    -d '{"name": "Bob", "avatar": "avatar_url"}'
  ```
- **Error Responses:**
  - `400 Bad Request`: Missing required fields.

---

### `GET /devices`
List all registered devices.

- **Description:** Returns an array of all known devices. Used by clients to populate their contact list on page load.
- **Response Body:**
  ```json
  [
    {
      "id": "dev-12345",
      "name": "Bob's iPhone",
      "avatar": "https://example.com/avatar.png",
      "created_at": "2026-10-01T10:00:00Z"
    }
  ]
  ```
- **Example cURL:**
  ```bash
  curl http://localhost:8080/devices
  ```

---

### `GET /devices/:id`
Get details of a single device.

- **Description:** Retrieves the profile of a specific device by its ID.
- **Response Body:**
  ```json
  {
    "id": "dev-12345",
    "name": "Bob's iPhone",
    "avatar": "https://example.com/avatar.png",
    "created_at": "2026-10-01T10:00:00Z"
  }
  ```
- **Example cURL:**
  ```bash
  curl http://localhost:8080/devices/dev-12345
  ```
- **Error Responses:**
  - `404 Not Found`: Device ID does not exist.

---

### `GET /health`
System health check.

- **Description:** Returns the current status and metrics of the gateway server.
- **Response Body:**
  ```json
  {
    "gateway_id": "gw-ny-01",
    "uptime_seconds": 345600,
    "connection_count": 1024
  }
  ```
- **Example cURL:**
  ```bash
  curl http://localhost:8080/health
  ```

---

### `GET /admin/events`
Server-Sent Events (SSE) stream for God-View dashboard.

- **Description:** Provides a real-time stream of system-wide events for monitoring purposes.
- **Event Types:**
  - `message_queued`: Message added to a mailbox.
  - `device_online`: Device connected.
  - `device_offline`: Device disconnected.
  - `delivery_acked`: Device acknowledged delivery.
  - `mailbox_depth`: Periodic update on total pending messages per device.
- **Response:** `text/event-stream` stream.
- **Example cURL:**
  ```bash
  curl -N http://localhost:8080/admin/events
  ```

---

### `GET /admin/mailbox/:device_id`
Check mailbox depth.

- **Description:** Returns the current depth and a preview of pending messages for a specific device, used by the God-View dashboard.
- **Response Body:**
  ```json
  {
    "device_id": "dev-12345",
    "depth": 4,
    "pending_messages": [
      {
        "message_id": "msg-001",
        "from": "dev-777",
        "queued_at": "2026-10-01T10:05:00Z"
      }
    ]
  }
  ```
- **Example cURL:**
  ```bash
  curl http://localhost:8080/admin/mailbox/dev-12345
  ```
- **Error Responses:**
  - `404 Not Found`: Device ID does not exist.

---

## WebSocket API

### `GET /ws?device_id=<id>`
Upgrade connection to WebSocket.

- **Description:** The primary endpoint for realtime messaging. Clients connect and upgrade the HTTP request to a persistent WebSocket connection. The `device_id` parameter uniquely identifies the connecting client.
- **Handshake:** After a successful HTTP upgrade to `101 Switching Protocols`, the connection operates entirely using the Realtime Protocol. The client must immediately send a `hello` frame to initiate state synchronization.
- **Reference:** For details on frames, payloads, and the message state machine, see [REALTIME_PROTOCOL.md](./REALTIME_PROTOCOL.md).
