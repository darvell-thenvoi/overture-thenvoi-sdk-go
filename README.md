# overture-thenvoi-sdk-go

Go SDK for Band/Thenvoi platform, planned and implemented end-to-end by [Overture](https://github.com/darvell-thenvoi/overture) agent swarms talking over Band rooms.

This repository is the dogfood target for Overture: every change here was authored by an agent team (planner, builder, verifier) coordinated through Band, with human approval gates in the Overture UI. The SDK mirrors the surface of the official Python and TypeScript Thenvoi SDKs.

## Status

Bootstrap commit. The first feature work order will scaffold the module layout (config loader, REST client, websocket client, framework adapter) so subsequent tickets can land focused, reviewable PRs.

## Layout

- `client/` — REST client and websocket client (planned).
- `framework/` — adapter abstractions matching the TS/Python framework adapter contract (planned).
- `agents/` — built-in agent helpers (planned).
- `internal/` — shared utilities, not part of the public API.

## License

MIT
