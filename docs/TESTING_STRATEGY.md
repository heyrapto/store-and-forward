# Testing Strategy

## 1. Philosophy
Test at the use-case layer with fakes, not mocks. This ensures tests are decoupled from implementation details. Integration tests should be used for adapters. No end-to-end tests in CI (too slow/flaky), but a manual demo script should be maintained.

## 2. Unit Tests
Use cases should be tested with in-memory fakes (e.g., `FakeDeviceRepo`, `FakePresenceStore`).
- **Examples:**
  - `send_message` to an offline device → delivery row is pending.
  - `ack_delivery` → row is flipped, receipt is emitted.
  - `sweep_expired` → rows older than the TTL are removed.

## 3. Integration Tests
SQLite repositories must be tested against a real temporary database. The WebSocket handler should be tested with a real HTTP test server. Crucially, test that the mailbox drains in seq order on reconnect.

## 4. Fake Implementations
Every port interface must have a fake implementation located in `internal/usecase/port/fake/`. These fakes must be deterministic and inspectable (e.g., `FakePresenceStore.IsOnline(id)` returns exactly what was explicitly set).

## 5. Load Tests
A Go bot swarm located in `server/test/loadbot/` will spawn N goroutines, each holding a WebSocket connection, sending messages, and acking deliveries.
- **Measures:** Message latency p50/p95/p99, mailbox drain time on reconnect, memory, and goroutine count under load.

## 6. Coverage Targets
- **Domain:** 100%
- **Use Cases:** 100%
- **Adapters:** Integration tested
- **Transport:** Integration tested
- **Frontend:** Unit tests for state logic (Zustand stores), component tests for critical flows.

## 7. Test File Locations
Test files map to the directory structure. `*_test.go` files sit alongside the code they test.
