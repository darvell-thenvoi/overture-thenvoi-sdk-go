# band-sdk-go

Go SDK for the Band platform, planned and implemented end-to-end by [Overture](https://github.com/darvell-thenvoi/overture) agent swarms talking over Band rooms.

The Go module path intentionally remains `github.com/darvell-thenvoi/overture-thenvoi-sdk-go` for this compatibility release; only runtime defaults and public branding move to Band here.

This repository is the dogfood target for Overture: every change here was authored by an agent team coordinated through Band, with human approval gates in the Overture UI. The SDK mirrors the agent REST surface of the Band TypeScript SDK generated from `@thenvoi/rest-client@0.0.113`.

## Status

REST client v1 with generated-contract-aligned helpers for agent identity, chat rooms, text messages, chat events, message processing status, participants, context, peers, contacts, and memories.

## REST client

```go
package main

import (
	"context"
	"log"

	"github.com/darvell-thenvoi/overture-thenvoi-sdk-go/client"
)

func main() {
	sdk := client.New(client.WithApiKey("band_agent_api_key"))

	agent, err := sdk.GetAgent(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	log.Println(agent.ID)
}
```

The default base URL is `https://app.band.ai`. Requests send `X-API-Key`, `Accept: application/json`, and a default SDK `User-Agent`.

## Agent REST helpers

- `GetAgent` / `GetAgentMe`
- `CreateChatRoom`, `GetChatRoom`, `ListChatRooms`
- `SendChatMessage`, `CreateChatEvent`, `ListChatMessages`, `GetNextChatMessage`
- `MarkChatMessageProcessing`, `MarkChatMessageProcessed`, `MarkChatMessageFailed`
- `ListChatParticipants`, `AddChatParticipant`, `RemoveChatParticipant`
- `GetChatContext`
- `ListPeers`
- `ListContacts`, `AddContact`, `RemoveContact`, `ListContactRequests`, `RespondContactRequest`
- `ListMemories`, `CreateMemory`, `GetMemory`, `SupersedeMemory`, `ArchiveMemory`

## Layout

- `client/` — REST client and agent API helpers.
- `framework/` — adapter abstractions matching the TS/Python framework adapter contract (planned).
- `agents/` — built-in agent helpers (planned).
- `internal/` — shared utilities, not part of the public API.

## License

MIT
