# queuepost-engine

## Domain Architecture & Core Purpose

This repository houses a **Multi-Platform Social Media Scheduler & Publisher** designed to automate content distribution across social channels. 

### Key Subsystems
1. **OAuth 2.0 Integration**: Handles secure connections to external social APIs (LinkedIn, Meta/Instagram) with automated token-refresh routines.
2. **Post & Media Management**: Manages draft creation, media uploads, and metadata persistence inside PostgreSQL.
3. **Async Scheduler Engine**: Utilizes Redis and Go `asynq` worker queues to pick up tasks precisely at the `scheduled_at` timestamp and dispatch API requests concurrently.
4. **Dashboard UI**: Provides a calendar-based scheduling interface with live preview cards and real-time execution status tracking.

## Better Fullstack Stack Parts & Ownership

`bts.jsonc` owns role selection and `ownerPartId` bindings.

- `backend.logging:go:zap` -> belongs to `backend:go:gin` (`apps/server`)
- `backend:go:gin` -> generated target `apps/server`
- `database:universal:postgres` -> generated target `packages/db`
- `frontend.animation:typescript:framer-motion` -> belongs to `frontend:typescript:react-vite`
- `frontend.css:typescript:tailwind` -> belongs to `frontend:typescript:react-vite`
- `frontend.forms:typescript:react-hook-form` -> belongs to `frontend:typescript:react-vite`
- `frontend.ui:typescript:nextui` -> belongs to `frontend:typescript:react-vite`
- `frontend:typescript:react-vite` -> generated target `apps/web`
- `workspaceRunner:universal:turborepo` -> root workspace

## Maintenance Guidelines for AI Agents

Keep `AGENTS.md` and `CLAUDE.md` updated when:
- Adding or removing workspace dependencies
- Changing project folder structures
- Introducing new backend services or frontend features
- Modifying build scripts or workflow execution commands
