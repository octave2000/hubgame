# Web Store Progress

## 2026-03-04
- Bootstrapped React + TypeScript + Vite frontend in `web/`.
- Integrated Tailwind CSS with custom cozy visual theme.
- Built advanced store UI with image-first cards, featured hero, and quick-view panel.
- Implemented advanced lookup: keyword search, category filter, mode filter, free/install toggles, sort options.
- Added keyboard shortcut (`Ctrl/⌘ + K`) to focus store search.
- Added responsive layout for desktop and mobile.
- Migrated package workflow to Bun (`bun.lock`, Bun scripts/docs).
- Redesigned UI to a calmer cozy look with larger game imagery, reduced initial text, no pricing labels, and softer lookup panel.
- Connected store catalog to live gateway API (`:8080`) with dev token bootstrap and websocket-based realtime refresh.
- Added documented backend seed flow so the live catalog can be populated with a versioned one-off command.
