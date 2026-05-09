package client

import (
	"context"
	"net/http"
	"time"
)

// ChatRoom describes a chat room returned by the platform.
type ChatRoom struct {
	ID         string    `json:"id"`
	InsertedAt time.Time `json:"inserted_at"`
	UpdatedAt  time.Time `json:"updated_at"`
	TaskID     *string   `json:"task_id"`
	Title      *string   `json:"title"`
}

// ChatRoomsMetadata describes pagination metadata for chat room listings.
type ChatRoomsMetadata struct {
	Page       *int `json:"page"`
	PageSize   *int `json:"page_size"`
	TotalCount *int `json:"total_count"`
	TotalPages *int `json:"total_pages"`
}

// ListChatRoomsResponse contains the authenticated agent's chat rooms.
type ListChatRoomsResponse struct {
	Data     []ChatRoom        `json:"data"`
	Metadata ChatRoomsMetadata `json:"metadata"`
}

// ListChatRooms fetches chat rooms available to the authenticated agent.
func (client *Client) ListChatRooms(ctx context.Context) (*ListChatRoomsResponse, error) {
	var out ListChatRoomsResponse
	if err := client.Do(ctx, http.MethodGet, "/api/v1/agent/chats", nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
