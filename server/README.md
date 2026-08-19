# Loremetry API

Account, credits, and AI analysis proxy for the Loremetry desktop app. Novel files stay on the writer’s machine.

```bash
go run ./cmd/server
```

Listens on `:5000` locally (or `PORT` on Miget).

## Local development

1. Start this server.
2. In the desktop app, use **Connect for AI** on the sign-in screen (token is saved in `loremetry-app.db`).
3. Dev token: `dev-token` (in-memory or Postgres seed user with Writer plan and 100 credits).

Empty token `empty-token` returns 402 on AI runs.

## Fireworks (serverless)

Calls go **direct** to Fireworks — no LangWatch, no TokenMix.

```bash
LOREMETRY_MODEL_BASE_URL=https://api.fireworks.ai/inference/v1
FIREWORKS_API_KEY=<key>
LOREMETRY_MODEL_FAST=accounts/fireworks/models/llama-v3p1-8b-instruct
LOREMETRY_MODEL_STRONG=accounts/fireworks/models/llama-v3p3-70b-instruct
```

`FIREWORKS_API_KEY` is an alias for `LOREMETRY_MODEL_API_KEY`. Without a key, runs use a stub completion.

## Postgres (Miget)

Set `DATABASE_URL` for persistent users, credits, job queue, and analysis cache.

```bash
DATABASE_URL=postgres://...
LOREMETRY_DEV_SEED=1   # recreates dev-token / empty-token users on migrate
```

## Job API

| Method | Path                  | Purpose                                                    |
| ------ | --------------------- | ---------------------------------------------------------- |
| POST   | `/analysis/jobs`      | Enqueue analysis (checks cache, credits, concurrent limit) |
| GET    | `/analysis/jobs/{id}` | Poll job status and result                                 |
| POST   | `/analysis/run`       | Deprecated shim — enqueue and wait                         |

Worker tuning:

```bash
LOREMETRY_AI_WORKERS_MIN=1
LOREMETRY_AI_WORKERS_MAX=4
```

Billing is an interface. Default is the fake in-memory provider. Swap `internal/apiserver/billing` when you pick Stripe or another vendor.
