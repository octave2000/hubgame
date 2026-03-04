# HubGame Web Store (Bun)

Frontend store UI for HubGame with live backend integration.

## Stack
- React + TypeScript + Vite
- Tailwind CSS v4
- Bun package/runtime workflow

## Dev Setup (with backend)
1. Start backend stack:
```bash
docker compose up --build
```
2. Seed starter catalog (one-off):
```bash
docker compose run --rm --profile tools seed-catalog
```
3. Generate static fallback catalog from `/games` (for offline/backend-down mode):
```bash
bun scripts/publish-games.mjs
bun scripts/sync-games-to-web.mjs
```
4. Start web app:
```bash
cd web
bun install
bun run dev
```

The app connects to gateway `http://localhost:8080` by default.

## Docker Compose Web Store
Run full stack including web store:
```bash
docker compose up --build
```

Web store URL:
- `http://localhost:3000`

## Optional Env
Create `web/.env`:
```bash
VITE_GATEWAY_URL=http://localhost:8080
VITE_DEV_TENANT_ID=hubgame-dev
VITE_DEV_USER_ID=web-dev-user
VITE_DEV_ROLE=developer
```

## Notes
- UI requests a dev token from gateway endpoint `POST /v1/auth/dev-token`.
- Catalog is loaded from backend entities with `kind=game`.
- Realtime updates use websocket stream on `topic=entity.game`.
- Leaderboards are loaded from backend (`/v1/leaderboard`) with global and per-game views.
- If gateway is unavailable, UI falls back to `web/public/fallback-catalog.json`.

## Build
```bash
bun run build
```

## Lint
```bash
bun run lint
```
