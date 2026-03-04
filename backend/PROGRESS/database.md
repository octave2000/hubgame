# Database Progress

## 2026-03-04
- Added SQLite schema initialization (entities + events).
- Implemented CRUD operations with soft-delete and version increment.
- Added optimistic concurrency checks with expected version handling.
- Added append-only event storage and topic-based pub/sub broker.
- Added schema-validation controller hook for entity and event payloads.
- Auto-emitted entity lifecycle events: inserted, updated, deleted.
- Added `seed_history` tracking table with `IsSeedApplied` / `MarkSeedApplied` helpers for versioned seed runs.
- Added leaderboard storage models on top of entities (`leaderboard_user`, `leaderboard_score`) and aggregation queries for global/per-game ranking.
