# overture-thenvoi-sdk-go

Go SDK for Band/Thenvoi platform, planned and implemented end-to-end by [Overture](https://github.com/darvell-thenvoi/overture) agent swarms talking over Band rooms.

This repository is the dogfood target for Overture: every change here was authored by an agent team (planner, builder, verifier) coordinated through Band, with human approval gates in the Overture UI. The SDK mirrors the surface of the official Python and TypeScript Thenvoi SDKs.

## Status

Bootstrap SDK with the REST client foundation in place. Endpoint-specific helpers, websocket support, and framework adapters will land in later focused PRs.

## REST client

```go
package main

import (
	"context"
	"log"
	"net/http"

	"github.com/darvell-thenvoi/overture-thenvoi-sdk-go/client"
)

func main() {
	sdk := client.New(client.WithApiKey("thenvoi_api_key"))

	var out map[string]any
	if err := sdk.Do(context.Background(), http.MethodGet, "/v1/agents/me", nil, &out); err != nil {
		log.Fatal(err)
	}
}
```

The default base URL is `https://platform.dev.thenvoi.com`. Requests send `Authorization: Bearer <api key>`, `Accept: application/json`, and a default SDK `User-Agent`.

## Layout

- `client/` — REST client foundation.
- `framework/` — adapter abstractions matching the TS/Python framework adapter contract (planned).
- `agents/` — built-in agent helpers (planned).
- `internal/` — shared utilities, not part of the public API.

## License

MIT
