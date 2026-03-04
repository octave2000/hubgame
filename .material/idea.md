# Open-Source Game Center Platform — Full Concept Document

## 1) Vision
Build an open-source **Game Center** where players can discover and play many games (solo, online multiplayer, offline multiplayer, and human vs bots), and where developers can add new games easily.

The platform is powered by a **Go-based backend** that combines:
- **A scalable real-time game server** (WebSocket-first)
- **A custom database + event system** designed for heavy load and real-time state
- **A secure controller layer** for authentication, encryption, and multi-tenant data governance

Frontend games can be written in any language/framework (Next.js, React, etc.) and run inside a **PWA-based game center shell**. Players can “install” individual games (pin to home screen) while still benefiting from shared features like chat, voice, and reactions.

---

## 2) Key Principles
1. **Open-source platform**: core server, SDKs, CLI, and game center shell are open.
2. **No monetization**: free to use.
3. **Extensible by design**: games plug into shared real-time features and event mechanics.
4. **Event-driven gameplay**: server logic is driven by structured events (send/receive/test).
5. **Secure multi-tenant**: the database engine is open-source, but user data is protected by a controller with auth + encryption.
6. **Container-first**: everything runs in containers; no external DB/services required.

---

## 3) Product Overview
### 3.1 Player Experience
- Browse a curated store/list of games (maintained by hubfly.cloud).
- Play instantly in the game center.
- Optional: **Install** a game as a PWA (pin to home screen) for quick launch.
- Every game can include:
  - In-game **chat** (text)
  - **Voice call** (real-time audio)
  - Emoji/reaction picker (quick expressions during play)
  - Presence indicators (online, in match, idle)

### 3.2 Developer Experience
Developers can create games by:
- Building the game UI with a **HubGame Center SDK**
- Implementing required hooks/events
- Using a **CLI** to scaffold, test, and publish
- Submitting their game by registering a repo in the store (you review + approve + merge)

They get:
- A **test connection environment** for event testing anytime
- Tools to simulate players, events, latency, retries
- Built-in shared features (chat/voice/reactions) without implementing them from scratch

---

## 4) Architecture (High Level)
### 4.1 Components
1. **Game Center Shell (Frontend)**
   - Store UI
   - Game launcher
   - PWA install support
   - Shared UI modules: chat, voice, emoji/reactions, player profiles

2. **Game SDK (Frontend)**
   - Connect/auth
   - Join match / create room
   - Send/receive events
   - Subscribe to state updates
   - Access shared features (chat/voice/reactions/presence)

3. **Real-Time Gateway (Go)**
   - WebSocket server (primary)
   - Optional: HTTP API for non-real-time operations
   - Handles:
     - Auth/session
     - Rate limiting
     - Topic subscriptions
     - Match routing

4. **Game Orchestrator (Go)**
   - Manages rooms, matches, lobbies, and players
   - Runs matchmaking
   - Applies rules and validates events
   - Coordinates bots

5. **Custom Database / Event Store (Go)**
   - Stores state + event log
   - Optimized for real-time reads/writes
   - Supports scaling and heavy load

6. **Controller / Security Layer (Go)**
   - Authentication + authorization
   - Multi-tenant boundaries
   - Encryption for sensitive data
   - Key management

7. **Voice/RTC Service (Containerized)**
   - WebRTC SFU/MCU or compatible audio service
   - Integrated with auth/presence

---

## 5) Core Concepts

### 5.1 Event-Driven Gameplay
Everything in a game is expressed as an event.

Examples:
- `match.create`
- `match.join`
- `move.place`
- `turn.end`
- `modifier.apply`
- `chat.send`
- `reaction.send`

Server responsibilities:
- Validate events (auth, permissions, schema)
- Apply game rules (server authoritative)
- Persist to event store
- Broadcast updates to subscribers

### 5.2 Game State Model
Each match has:
- **Snapshot state** (current board / timers / inventories)
- **Event log** (append-only history)
- **Derived views** (leaderboards, stats, match summaries)

To scale:
- Snapshot periodically
- Replay events for recovery
- Cache hot state in memory with durable persistence

### 5.3 Rooms, Lobbies, Matches
- **Lobby**: waiting area, invites, matchmaking
- **Room**: group context (chat/voice)
- **Match**: the actual game instance (rules + state)

---

## 6) Game Modes (Platform-Level)
Each game can expose modes like:
- Solo
- Human vs Bot
- Offline local multiplayer
- Online multiplayer
- Ranked / Unranked
- Custom rulesets
- Party mode

The platform provides common building blocks:
- Turn timers
- Rematch
- Spectator mode
- Reconnect/resume
- Anti-cheat basics (rate limits, authoritative rules)

---

## 7) First Game: Tic-Tac-Toe (Advanced)
### 7.1 Baseline
- Classic 3x3 and optional larger boards (4x4, 5x5)
- Win conditions configurable (3-in-row, 4-in-row)

### 7.2 “Modifiers” System (Crazy Modes)
Modifiers are power-ups that can alter flow.

Examples:
- Skip opponent turn
- Take two moves
- Undo last move (with constraints)
- Freeze a cell
- Swap symbols
- Block a row/column temporarily

How modifiers are obtained:
- Earned via gameplay achievements
- Collected/owned by player inventory
- Dropped randomly in special modes
- Granted by the mode itself (everyone gets a random modifier)

Server must:
- Validate modifier rules
- Ensure fairness (cooldowns, caps)
- Persist inventory changes

---

## 8) Shared Social Features
### 8.1 Chat
- Game room chat, match chat
- Emoji quick picker
- Moderation hooks (spam limits, reports)

### 8.2 Reactions
- Instant emoji reactions (non-spammy)
- Optional reaction feed overlay

### 8.3 Voice Call
- Room-based voice
- Push-to-talk optional
- Integrated auth tokens

### 8.4 Presence
- Online/offline
- In lobby/in match
- In voice

---

## 9) Publishing Workflow
1. Developer creates a game via CLI scaffold.
2. They implement required manifest + event handlers + UI.
3. They run tests locally against a provided test environment.
4. They publish by:
   - Adding repo URL + metadata in the store submission form
   - Pushing code to their repo
5. hubfly.cloud review, merge/approve into the store.
6. The store index updates and the game becomes available.

Guiding rules:
- Store is curated (quality + security)
- Games must not require external services by default(like another side database)
- Everything should run containerized

---

## 10) Security Model
### 10.1 Authentication
- Session tokens for clients
- Signed tokens for WebSocket connections

### 10.2 Authorization
- Game-level permissions (who can create matches, join rooms)
- Role-based admin controls

### 10.3 Data Encryption
- Encrypt sensitive records (e.g., private user metadata, tokens)
- Support per-tenant encryption keys

### 10.4 Anti-Abuse Baselines
- Rate limiting on events
- Schema validation
- Replay protection
- Flood control for chat/reactions

---

## 11) Scalability Plan
### 11.1 Horizontal Scaling
- Stateless gateway nodes
- Sharded match routing
- Partitioned event store

### 11.2 Hot State
- Keep active matches in memory for low latency
- Persist events asynchronously (with durability guarantees)

### 11.3 Load Patterns
- Many concurrent sockets
- Bursty events (moves, reactions)
- Voice traffic isolated in its own service

---

## 12) Developer Tooling
### 12.1 CLI Capabilities
- `create-game` scaffold
- `dev` local runner
- `test` simulate events
- `validate` manifest + schema
- `publish` submission helper
- `lint` best practices

### 12.2 Game Manifest
Each game includes:
- Name, icon, screenshots
- Supported modes
- Required permissions
- Event schemas
- Versioning

### 12.3 SDK
Provides:
- Auth + connect
- Room/match helpers
- Event send/subscribe
- State sync helpers
- Social features integration

---

## 13) Container Deployment
### 13.1 Services (Example)
- `gateway` (WS + HTTP)
- `orchestrator`
- `db-engine`
- `controller`
- `voice-service`
- `web-shell` (game center)

### 13.2 Operational Needs
- Metrics + logs
- Health checks
- Rolling updates
- Backups for event store

---

## 14) Governance & Contribution
- Core platform is open-source.
- Games are open-source, contributed via repos.
- You maintain the store index + review/approval.

Contribution guidelines:
- Clear code standards
- Security checks
- Game must use platform SDK
- No hidden external dependencies

---

## 15) Suggested Refinements & Extra Ideas
### 15.1 Game Categories
- Casual
- Strategy
- Party
- Puzzle
- Educational

### 15.2 Player Profile System
- Achievements
- Inventory (modifiers, cosmetics)
- Match history

### 15.3 Spectator & Streaming Mode
- Watch live matches
- Delay mode
- Chat-only spectator

### 15.4 Offline-First PWA Behavior
- Cache game assets
- Allow bot/solo modes offline
- Sync results when online

### 15.5 Plugin System for Shared Features
- Swap emoji picker UI
- Optional mini-games in chat
- Event transformers for testing

---

## 16) Milestone Roadmap (Practical Order)
1. **MVP Platform**
   - Gateway + basic auth
   - Room + match creation
   - Event send/receive
   - Minimal store UI

2. **Tic-Tac-Toe v1**
   - Online 1v1
   - Classic rules

3. **Tic-Tac-Toe Modifiers**
   - Inventory + modifier events
   - Balancing rules

4. **Shared Chat + Reactions**
   - Room chat
   - Reaction overlay

5. **Bots Framework**
   - Bot adapters + difficulty

6. **Voice**
   - WebRTC voice rooms

7. **Developer CLI + SDK Polish**
   - Templates
   - Test environment

8. **Scaling & Reliability**
   - Sharding
   - Snapshots
   - Monitoring

---

## 17) Glossary
- **Event**: a structured message representing an action.
- **Snapshot**: compact state image of a match.
- **Event Log**: append-only history enabling replay/recovery.
- **Game SDK**: client library to integrate games with platform.
- **Controller**: security layer for data governance and auth.

---

## 18) One-Sentence Summary
An open-source, containerized, Go-powered game center where any developer can ship games using an event-driven real-time backend with built-in chat/voice/reactions, starting with an advanced tic-tac-toe featuring collectible gameplay modifiers.

---

## 19) Database Backbone (Go) — Detailed Blueprint

### 19.1 Position in HubGame
The database backbone is not only a persistence layer. It is a **state + event + realtime core** with embedded controllers for policy enforcement.

Responsibilities:
- Durable storage (primary: SQLite-compatible mode, future pluggable engines)
- Event sourcing (append-only timeline for state reconstruction)
- Live subscriptions (native websocket fanout)
- Embedded lifecycle controllers (auth, tenant rules, validation, anti-abuse)

### 19.2 Storage Engine Model
1. **Entity Store**
- Multi-tenant records (`tenant_id`, `id`, `kind`, `data`, `version`, timestamps)
- Soft delete + optimistic versioning
- Secondary indexes by tenant/kind/update time

2. **Event Store**
- Append-only events (`topic`, `key_ref`, `type`, `payload`, `created_at`)
- Ordered IDs for reliable replay
- Per-topic scan for catch-up sync

3. **Snapshot Layer (Phase 2)**
- Periodic compressed snapshots for active matches/rooms
- Replay from nearest snapshot + forward events
- Fast warm-recovery on node restart

4. **In-Memory Hot Cache (Phase 2/3)**
- Track active match state in memory
- Async flush + durability guarantees
- Write-back strategy with bounded lag and health alarms

### 19.3 Embedded Controller Pipeline
Database operations pass through controller hooks:
- `BeforeInsert`
- `BeforeUpdate`
- `BeforeDelete`
- `BeforeAppendEvent`

Controllers are composable and ordered, enabling reusable cross-cutting logic:
- Tenant boundary enforcement
- Role-based authorization checks
- Payload schema validation
- Rate-limits and flood protection
- Audit metadata injection

This creates a **policy-aware database**, not just CRUD storage.

### 19.4 Native Realtime Streaming
- Topic-based pub/sub from event store writes
- WebSocket endpoint for low-latency subscriptions
- Backpressure-safe broadcast semantics (drop policy + metrics)
- Historical replay (`after_id`) + live tail pattern

### 19.5 Planned Query and Write Features
- Atomic transactions with batch operations
- Conditional update with expected version (`CAS` semantics)
- Upsert support for selected entity kinds
- Range queries for timeline and analytics projections
- TTL/retention policies for transient data

### 19.6 SQLite-First, Engine-Flexible
Initial runtime mode:
- SQLite WAL mode for local durability and high concurrent reads
- Busy timeout + tuned pragmas

Future adapters:
- PostgreSQL adapter for large distributed workloads
- Embedded LSM mode for write-heavy telemetry/event channels

The API surface remains stable through a storage interface.

---

## 20) Controller Backbone (Go) — Detailed Blueprint

### 20.1 Core Scope
Controller layer is the trust and governance engine:
- Authentication (JWT/session flows)
- Authorization (tenant + role + action)
- Security policy execution inside DB hooks
- Token tooling for API, WebSocket, CLI, bots

### 20.2 Auth Model
- Short-lived access token for API/WebSocket
- Refresh token rotation (Phase 2)
- Issuer/audience enforcement
- Service account tokens for automation

### 20.3 Authorization Model
- Tenant isolation as hard baseline
- Role matrix:
  - `player`
  - `moderator`
  - `developer`
  - `tenant_admin`
  - `platform_admin`
- Action-scoped permissions (`match.create`, `room.moderate`, `bot.run`)

### 20.4 Security Add-ons
- Event signing/verification for anti-tamper traces
- Replay attack protection via nonce + expiry windows
- Optional field-level encryption for sensitive user attributes
- Audit trail streams for security operations

---

## 21) Testing and Reliability Strategy

### 21.1 Database and Controller Test Matrix
- Unit: CRUD/versioning/hook enforcement
- Integration: auth middleware + DB hooks + websocket stream
- Load: concurrent writes + fanout + replay scans
- Chaos: abrupt restart + recovery from event log/snapshot

### 21.2 Reliability Controls
- Health checks per component
- Migration version table + startup compatibility check
- Structured logs with request/tenant correlation IDs
- Metrics:
  - write latency p50/p95/p99
  - subscriber lag
  - dropped events
  - auth failure rate

---

## 22) Delivery Roadmap for Backbone (Execution)

1. **Phase A: Foundation (Now)**
- SQLite entity/event store
- CRUD + append event APIs
- WebSocket topic stream
- JWT auth + tenant controller hooks

2. **Phase B: Trust and Correctness**
- RBAC action matrix
- Schema validation controller
- expected-version conditional updates
- audit logging controller

3. **Phase C: Performance and Scale**
- snapshots + replay APIs
- batch writes + transaction helpers
- cache layer for hot matches
- benchmark and profiling suite

4. **Phase D: Production Hardening**
- retention/compaction jobs
- dead-letter event handling
- backup/restore workflows
- operational dashboards and alerting

---

## 23) Competitive/Unique Platform Advantages
- **Controller-embedded database** for policy-aware persistence.
- **Realtime-first event core** with replay + live tail in one model.
- **Game-focused primitives** (rooms/matches/modifiers/social events) on top of generic storage.
- **Container-native deployment** with minimal external dependencies.
- **Open-source extensibility** for games, controllers, and tooling.
