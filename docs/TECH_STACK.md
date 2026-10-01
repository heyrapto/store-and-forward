# Tech Stack Reference

## Backend

| Component | Choice | Why it was chosen | Alternative Considered | Link / Import Path |
| :--- | :--- | :--- | :--- | :--- |
| **Language** | Go 1.22+ | Excellent concurrency (goroutines) for WS connections, strong stdlib, high performance. | Node.js, Rust | `go` |
| **HTTP Router** | stdlib `net/http` mux | Go 1.22 introduced enhanced routing in the stdlib, eliminating the need for third-party routers for basic paths. | Chi, Gorilla Mux | `net/http` |
| **WebSocket** | `coder/websocket` | Modern, actively maintained, minimal, and context-aware WS library. | `gorilla/websocket` (archived/unmaintained) | `github.com/coder/websocket` |
| **Database** | `modernc.org/sqlite` | Pure Go SQLite driver. No CGO required, simplifying cross-compilation and docker builds. Native WAL mode support. | `mattn/go-sqlite3` (requires CGO) | `modernc.org/sqlite` |
| **Migrations** | `goose` | Simple, robust, SQL-based migration tool that integrates natively with Go `database/sql`. | `golang-migrate` | `github.com/pressly/goose/v3` |
| **KV / PubSub** | `go-redis/v9` | Standard, highly performant Redis client for Go. Essential for cross-node presence and pub/sub. | `redigo` | `github.com/redis/go-redis/v9` |
| **Logging** | `slog` | Go 1.21+ standard library structured logger. Fast and unified. | `zap`, `logrus` | `log/slog` |
| **IDs** | `oklog/ulid` | Lexicographically sortable, collision-resistant unique identifiers. Better DB index performance than UUIDv4. | UUIDv4 | `github.com/oklog/ulid/v2` |

## Frontend

| Component | Choice | Why it was chosen | Alternative Considered | Link |
| :--- | :--- | :--- | :--- | :--- |
| **Framework** | React 18 | Industry standard, component-based, massive ecosystem for UI building. | Vue, Svelte | [react.dev](https://react.dev) |
| **Language** | TypeScript | Type safety across the wire, prevents runtime errors in state management. | JavaScript | [typescriptlang.org](https://www.typescriptlang.org/) |
| **Bundler** | Vite | Lightning-fast HMR, built-in TS support, significantly faster than Webpack. | Create React App | [vitejs.dev](https://vitejs.dev) |
| **State** | Zustand | Unopinionated, minimal boilerplate, extremely fast global state management. | Redux Toolkit | [zustand-demo.pmnd.rs](https://zustand-demo.pmnd.rs/) |
| **Styling** | Tailwind CSS v3 | Utility-first CSS. Enables rapid prototyping and styling without context switching. | CSS Modules, MUI | [tailwindcss.com](https://tailwindcss.com) |
| **Animations**| Framer Motion | Declarative animations for React. Perfect for fluid UI interactions (message bubbles, transitions). | CSS Transitions | [framer.com/motion](https://framer.com/motion) |
| **Local DB** | IndexedDB | Browser-native durable storage. Required for the client-side outbox and local chat history. | LocalStorage | [MDN IndexedDB](https://developer.mozilla.org/en-US/docs/Web/API/IndexedDB_API) |

## Infrastructure

| Component | Choice | Why it was chosen |
| :--- | :--- | :--- |
| **Containerization** | Docker Compose | Easiest way to spin up the multi-container environment (Gateways, Redis, Caddy, UI). |
| **Reverse Proxy** | Caddy | Automatic HTTPS, incredibly simple configuration (Caddyfile), built-in load balancing. |
| **Cache/Bus** | Redis 7 | Industry standard for transient state and lightweight Pub/Sub messaging. |
| **Storage** | SQLite | Serverless, single-file. Mounted via shared volume across gateway containers. Configured in WAL (Write-Ahead Logging) mode to allow concurrent readers/writers. |

## Testing

| Component | Choice | Why it was chosen |
| :--- | :--- | :--- |
| **Unit Testing** | Go stdlib `testing` | Standard, fast, runs natively without plugins. |
| **Assertions** | `testify` | Clean `assert.Equal` syntax reduces boilerplate `if err != nil` assertions in test code. |
| **Mocks/Fakes** | Fakes & Stubs | Hand-written interface implementations for Ports to ensure Use Cases are tested in pure isolation without heavy mocking frameworks. |
| **Integration** | Real SQLite + Docker | Testing against real DBs ensures SQL queries and constraints act as expected. |
| **Load Testing** | k6 / Custom Go Bot Swarm | k6 allows JS-scripted WS testing. A custom Go bot swarm can easily simulate thousands of persistent goroutine connections to saturate the gateway. |
