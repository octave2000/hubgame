# HubGame Games Directory

Each game lives under `/games/<game-id>` and must include a `manifest.json`.

## Manifest rules
Required fields:
- `id`
- `name`
- `version`
- `author`
- `description`
- `entry`
- `categories` (non-empty array)
- `modes` (non-empty array)
- `supports` (object)

In-house rule:
- if `inhouse: true`, then `author` must be `hubgame`.

## Publish index
Generate publish index for all valid games:

```bash
bun scripts/publish-games.mjs
```

Output file:
- `games/.published/index.json`

## Sync to Web Fallback Catalog
Copy game packages and generate web fallback file:

```bash
bun scripts/sync-games-to-web.mjs
```

Outputs:
- `web/public/games/*`
- `web/public/fallback-catalog.json`
