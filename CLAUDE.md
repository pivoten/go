# CLAUDE.md - pivoten/go

## THIS IS A PUBLIC REPOSITORY

`pivoten/go` is **public**. Everything committed here is world-readable, forever
(git history included). Treat every file - and every commit - as published.

## Hard rules (for humans and agents)

- **NEVER commit secrets.** No API keys, tokens, passwords, private keys, certificates,
  connection strings, `.env` values, customer/employee data, or internal hostnames/IPs.
  Code here reads configuration from the environment at runtime - this repo holds the code,
  never the values.
- **No proprietary business logic.** This module is only for generic, reusable
  infrastructure (Temporal base, shared utilities). Product and business logic live in the
  private product repos (e.g. `pivotenctl`, `tax-services`, `missioncontrol-*`), which may
  *import* this module.
- If a change would embed a secret, private datum, or proprietary logic, **stop** - move it
  to config/env or to a private repo instead.
- Do not add example values that are real. Use placeholders (`john.doe`, `1.2.3.4`,
  `<base64-cert>`).

## Contributions and write access

Commits come from the **Pivoten team only**. Public visibility grants read and fork, **not
write** - push/merge access is limited to Pivoten collaborators. External parties may open
PRs, but nothing merges without Pivoten review.

## What belongs here

Generic, dependency-light Go packages usable across Pivoten projects, each with tests and no
secrets.

- `temporal/` - Temporal client + worker bootstrap with mTLS and AES-256-GCM payload
  encryption, configured from environment variables.
