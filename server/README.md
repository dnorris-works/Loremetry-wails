# Loremetry API

Account, credits, and AI analysis proxy for the Loremetry desktop app. Novel files stay on the writer’s machine.

```
go run ./cmd/server
```

Listens on `:8080` locally (or `PORT` env on Miget).

## Local development

1. Start this server.
2. In the desktop app, use **Connect for AI** on the sign-in screen (token is saved in `loremetry-app.db`).
3. Dev token: `dev-token` (seeded user with a Writer plan and 100 credits).

Empty token `empty-token` returns 402 on AI runs.

Set `LOREMETRY_MODEL_BASE_URL` and `LOREMETRY_MODEL_API_KEY` to use a real OpenAI-compatible gateway. Without them, runs use a stub completion.

Billing is an interface. Default is the fake in-memory provider. Swap `internal/apiserver/billing` when you pick Stripe or another vendor.
