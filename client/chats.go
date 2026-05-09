package client

import (
	"context"
	"errors"
	"net/http"
	"net/url"
)

// ChatRoom describes a chat room returned by the platform.
type ChatRoom struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Description *string `json:"description"`
	Type        *string `json:"type,omitempty"`
	Status      *string `json:"status,omitempty"`
}

// GetChatRoom fetches a chat room by ID.
func (client *Client) GetChatRoom(ctx context.Context, chatID string) (*ChatRoom, error) {
	if chatID == "" {
		return nil, errors.New("thenvoi: chat id is required")
	}

	escapedChatID := url.PathEscape(chatID)

	var out struct {
		Data ChatRoom `json:"data"`
	}

	err := client.Do(ctx, http.MethodGet, "/api/v1/agent/chats/"+escapedChatID, nil, &out)
	if err != nil {
		return nil, err
	}

	return &out.Data, nil
}
