# 0005 – One VPS Behind Caddy (Demo Deployment)

**Status:** Accepted  
**Date:** 2026-10-01  
**Deciders:** Series author

## Context
To demonstrate the distributed nature of the messaging system, we need a deployment strategy that shows multiple gateway instances interacting with shared resources (Redis and SQLite). The deployment must be simple enough for learners to understand and replicate, while still demonstrating real-world concepts like load balancing, TLS termination, and WebSocket routing. 

## Decision
The demo deployment target will be a single Virtual Private Server (VPS) running Docker Compose. 

- The deployment will consist of 3 Go gateway instances.
- Caddy will be used as a reverse proxy and load balancer in front of the gateways.
- Caddy will handle TLS termination (via Let's Encrypt automatic HTTPS) and WebSocket upgrades.
- Caddy will use round-robin routing to distribute connections across the 3 gateway instances.
- Cloudflare will be used for DNS management in front of the server.
- The configuration files (`docker-compose.yml` and `Caddyfile`) will be committed to the `server/deploy/` directory.

### Rejected Alternatives
- **Kubernetes:** Introduces far too much operational ceremony and complexity for an educational demo focused on application architecture.
- **Nginx:** While capable, Caddy's automatic HTTPS configuration is significantly simpler to set up and demonstrate.

## Consequences
- **Educational Value:** Demonstrates how multiple gateways sharing Redis (for pub/sub and coordination) and SQLite (for durable storage) can handle transparent reconnection and mailbox drain, even when clients are routed to different instances.
- **Accessibility:** Docker Compose provides a low barrier to entry for learners to spin up the entire cluster locally or on a cheap VPS.
- **Scalability Limits:** A single VPS deployment is a vertical scaling limit, but it perfectly serves the purpose of demonstrating concurrent multi-instance behavior.
