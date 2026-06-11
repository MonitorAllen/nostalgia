# Deployment Health Checks

Nostalgia exposes two operational probes from the API service:

- `GET /healthz`: process liveness. It returns `200` with `{"status":"ok"}` when the Gin API process can serve HTTP.
- `GET /readyz`: dependency readiness. It returns `200` with `{"status":"ready"}` only when PostgreSQL and Redis both respond to `Ping`; otherwise it returns `503` with failing checks marked as `unavailable`.

Nginx proxies HTTPS `/healthz` and `/readyz` to the API service. The HTTP server keeps a local `/healthz` response for the web container healthcheck while all other HTTP traffic is redirected to HTTPS.

Docker Compose healthchecks:

- `postgres`: `pg_isready` against the configured database.
- `redis`: `redis-cli ping`.
- `api`: `GET http://localhost:8080/healthz`.
- `web`: `GET http://localhost/healthz`.

The API service waits for healthy PostgreSQL and Redis services. The web service waits for a healthy API service.
