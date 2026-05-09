package client

import (
	"context"
	"net/http"
)

// ChatRoom describes a chat room returned by the platform.
type ChatRoom struct {
	ID   string  `json:"id"`
	Name *string `json:"name"`
}

// ListChatRooms fetches chat rooms visible to the authenticated agent.
func (client *Client) ListChatRooms(ctx context.Context) ([]ChatRoom, error) {
	var out struct {
		Data []ChatRoom `json:"data"`
	}

	if err := client.Do(ctx, http.MethodGet, "/v1/chats", nil, &out); err != nil {
		return nil, err
	}

	return out.Data, nil
}
