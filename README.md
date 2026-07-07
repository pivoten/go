# pivoten/go

Shared Go packages for Pivoten services. Import as `github.com/pivoten/go/<pkg>`.

This module holds generic, reusable infrastructure with **no secrets and no
proprietary business logic**, so it can be consumed by any Pivoten project.

## Packages

### `temporal`

A Temporal client + worker bootstrap with **mTLS** and **AES-256-GCM payload
encryption**, configured from environment variables. It mirrors the conventions
in `missioncontrol-temporal-workers` so services can converge on one base.

- `Config` - env-driven config (`LoadConfig`), with `TLSConfig()` (mTLS from
  base64 cert material) and `DataConverter()` (encryption when a codec key is set).
- `NewClient` / `ConfigToClient` / `NewClientWithRetry` - dial + health-check.
- `WorkerBuilder` - register workflows/activities and run a worker.
- `Codec` - AES-256-GCM `PayloadCodec`; `NewEncryptionDataConverter` wraps it.

```go
cfg, _ := temporal.LoadConfig("")           // reads TEMPORAL_* env
logger, _ := zap.NewProduction()
c, err := temporal.ConfigToClient(ctx, cfg, logger)  // mTLS + encryption applied
defer c.Close()

temporal.NewWorkerBuilder(temporal.WorkerConfig{Client: c, TaskQueue: cfg.TaskQueue, Logger: logger}).
    RegisterWorkflows(MyWorkflow).
    RegisterActivities(MyActivities).
    Run()
```

#### Environment

| Var | Purpose |
|-----|---------|
| `TEMPORAL_ADDRESS` | host:port (default `localhost:7233`) |
| `TEMPORAL_NAMESPACE` | namespace (default `default`) |
| `TEMPORAL_TASK_QUEUE` | task queue name |
| `TEMPORAL_CLIENT_CERT_BASE64` / `TEMPORAL_CLIENT_KEY_BASE64` | base64 mTLS client cert/key |
| `TEMPORAL_SERVER_ROOT_CA_BASE64` / `TEMPORAL_SERVER_NAME` | base64 server root CA + SNI |
| `TEMPORAL_INSECURE_SKIP_VERIFY` | skip TLS verify (dev only) |
| `TEMPORAL_API_KEY` | API-key auth (alternative to mTLS) |
| `TEMPORAL_CODEC_KEY_BASE64` | 32-byte base64 key; enables payload encryption |
| `TEMPORAL_CODEC_KEY_ID` | key id recorded on payloads (default `pivoten-1`) |

Encrypted payloads carry `encoding: binary/encrypted`; decrypt them in the
Temporal Web UI with a codec server (`pivoten/temporal-codec-server`) that shares
the same key.

## Development

```sh
go test ./...
```

No secrets or credentials belong in this repo - it is intentionally public so any
project's CI can pull it without cross-repo auth.
