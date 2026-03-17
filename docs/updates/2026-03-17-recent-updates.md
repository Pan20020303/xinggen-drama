# 2026-03-17 Recent Updates

## Backend and Infrastructure

- Improved Volces ARK client reliability:
  - Added transient network retry for request failures (including TLS handshake timeout).
  - Added explicit HTTP transport tuning for timeout-sensitive calls.
  - Added tests for retry path and task-status usage parsing.
- Fixed video token accounting flow:
  - Record usage only when token usage is non-zero.
  - For async video tasks, record usage delta on completion to avoid double counting.
- Added MySQL migration/startup support for local deployment:
  - Added `scripts/start_mysql.sh` to start MySQL, migrate from SQLite, and start app.
  - Removed deprecated MySQL 8.4 startup flag from `docker-compose.yml`.
  - Hardened SQLite->MySQL migrator to sanitize invalid UTF-8 / null bytes.
  - Added MySQL AutoMigrate compatibility handling for ignorable `Error 1091` unique-index drop failures.

## Frontend

- Updated API typings and request adapters for assets, dramas, frames, generation, images, props, and videos.
- Updated editor and workflow views:
  - Timeline/Storyboard/Video editor related components and stores.
  - Episode workflow and professional editor interaction updates.
- Updated admin pages:
  - AI config, billing, token stats, users, and admin layout integration.
- Added/updated type declarations:
  - asset, drama, generation, image, timeline, video.

## Docs

- Added this update log to `docs/updates` for release traceability.
- Synced additional docs assets and planning artifacts under `docs/`.
