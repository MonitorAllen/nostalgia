# Cloudflare Nginx Deployment Design

## Context

Nostalgia currently deploys with `caddy`, `web`, and `api` containers. Caddy owns public `80/443`, automatic TLS, and reverse proxy routing. The `web` container is already an Nginx static frontend container, so keeping a separate Caddy edge container is redundant once the public domain is proxied through Cloudflare.

The admin frontend is currently mounted at `/admin`, with `/backend` redirected to `/admin`. The new deployment direction makes `/backend` the canonical owner-only management path and removes the historical redirect.

## Goals

- Remove the Caddy container from production and development Compose files.
- Make `web`/Nginx the only public ingress container for `80/443`.
- Use Cloudflare as the public edge, with Cloudflare Origin CA certificates mounted into Nginx.
- Keep backend services internal on the Docker network:
  - `/api/*` proxies to `api:8080`.
  - `/v1/*` proxies to `api:9091`.
- Keep uploaded resources served by Nginx at `/resources/*`.
- Change the admin frontend canonical route from `/admin` to `/backend`.
- Remove `/backend -> /admin` redirects and update user-facing docs to `/backend`.

## Non-Goals

- No backend API path changes.
- No database changes.
- No Docker image registry changes.
- No automated Cloudflare API management in this branch.
- No committing TLS private keys, Origin CA certificates, or real deployment secrets.
- No direct browser-trusted TLS support without Cloudflare proxy. The production HTTPS model assumes Cloudflare Full Strict.

## Deployment Architecture

Production request path:

```text
Browser -> Cloudflare -> web(Nginx) -> api
```

Nginx will expose:

- `80`: redirects to HTTPS for production.
- `443`: serves the SPA and reverse proxies API paths with the Cloudflare Origin CA certificate mounted from the host.

The expected host certificate layout is:

```text
./certs/cloudflare-origin.pem
./certs/cloudflare-origin.key
```

These files must stay untracked. Compose mounts `./certs` read-only into `/etc/nginx/certs`.

Development images generate a local self-signed certificate during image build so `docker-compose.dev.yaml` can run without Cloudflare certificate files. Production Compose still mounts the Cloudflare Origin CA files read-only.

## Nginx Routing

Nginx owns the behavior previously split between Caddy and Nginx:

- `/api/` and `/api/*` proxy to `http://api:8080`.
- `/v1/` and `/v1/*` proxy to `http://api:9091`.
- `/resources/*` serves files from `/usr/share/nginx/resources/`.
- `/backend/*` and all other frontend routes use Vue SPA fallback.
- Static hashed assets under `/assets/*` keep long immutable caching.

Nginx should forward standard proxy headers:

- `Host`
- `X-Real-IP`
- `X-Forwarded-For`
- `X-Forwarded-Proto`

It should also trust Cloudflare visitor IP headers through `CF-Connecting-IP`. Exact Cloudflare CIDR maintenance can be handled later; this branch prepares the header forwarding and keeps source IP behavior sane for a Cloudflare-proxied origin.

## Frontend Routing

The Vue router changes:

- `/admin/login` becomes `/backend/login`.
- `/admin` becomes `/backend`.
- Named routes remain unchanged (`adminLogin`, `adminArticles`, etc.) so component navigation code can stay stable.
- Admin auth redirects and setup completion route to named routes or `/backend/...` constants.

No compatibility redirect from `/admin` to `/backend` is added. This keeps the route vocabulary clean and matches the user-approved direction.

## Verification

Minimum verification:

```bash
cd web/frontend && bun test
cd web/frontend && bun run type-check
cd web/frontend && bun run build
docker compose config
docker compose -f docker-compose.dev.yaml config
docker run --rm -v "$PWD/web/nginx.conf:/etc/nginx/nginx.conf:ro" -v "$PWD/web/frontend/dist:/usr/share/nginx/html:ro" -v "$PWD/resources:/usr/share/nginx/resources:ro" -v "$PWD/certs:/etc/nginx/certs:ro" nginx:alpine nginx -t
```

The final `nginx -t` requires local test certificate files or real Cloudflare Origin CA files at the expected `./certs` paths.
