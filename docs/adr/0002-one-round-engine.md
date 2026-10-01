# 0002 – One Delivery Engine (No Separate Worker Service)

**Status:** Accepted  
**Date:** 2026-10-01  
**Deciders:** Series author

## Context
Message delivery involves receiving a message, persisting it, fanning it out to recipients, and pushing it via WebSockets or queuing it for offline users. We need to decide how to deploy the workers that process this background fan-out logic.

## Decision
All message delivery logic (persistence → fan-out → push/queue) will run inside the gateway process itself using goroutines and a worker pool. There will be no separate queue consumer or worker service in Episodes 1 and 2 of the series.
A `DeliveryEngine` (or an inline use-case call) runs the fan-out in a bounded goroutine pool within the same application instance.

**Tradeoffs & Abstractions:**
Horizontal fan-out across multiple gateway instances will eventually require Redis pub/sub (planned for Episode 3). However, because we follow clean architecture, the engine interacts with a port interface. When the time comes, a Redis adapter can be swapped in without changing the core delivery engine logic.

**Rejected Alternatives:**
*   Separate worker service (e.g., reading from Kafka/RabbitMQ): Rejected because it adds significant operational complexity too early in the learning process, distracting from the core concepts being established in the initial episodes.

## Consequences
*   Simplifies deployment and local development for the first two episodes.
*   The gateway process consumes more CPU and memory as it handles both WebSocket termination and background delivery tasks.
*   If the gateway process crashes, any messages held in memory for fan-out that haven't been persisted to a shared queue might be dropped (though persistence happens before fan-out, meaning the message isn't lost, just its real-time push attempt).
