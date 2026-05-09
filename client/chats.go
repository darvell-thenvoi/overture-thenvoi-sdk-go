package client

import (
	"context"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

// ChatRoom describes a chat room returned by the platform.
type ChatRoom struct {
	ID         string     `json:"id"`
	InsertedAt time.Time  `json:"inserted_at"`
	TaskID     *string    `json:"task_id"`
	Title      *string    `json:"title"`
	UpdatedAt  *time.Time `json:"updated_at"`
}

// ListChatRoomsParams configures optional pagination for ListChatRooms.
type ListChatRoomsParams struct {
	Page     *int
	PageSize *int
}

// ResponseMetadata describes API pagination metadata.
type ResponseMetadata struct {
	Page       int `json:"page"`
	PageSize   int `json:"page_size"`
	TotalCount int `json:"total_count"`
	TotalPages int `json:"total_pages"`
}

// ListChatRoomsResponse contains the chat room list and pagination metadata.
type ListChatRoomsResponse struct {
	Data     []ChatRoom       `json:"data"`
	Metadata ResponseMetadata `json:"metadata"`
}

// ListChatRooms fetches chat rooms visible to the authenticated agent.
func (client *Client) ListChatRooms(ctx context.Context, params *ListChatRoomsParams) (*ListChatRoomsResponse, error) {
	path := "/api/v1/agent/chats"
	if params != nil {
		query := url.Values{}
		if params.Page != nil {
			query.Set("page", strconv.Itoa(*params.Page))
		}
		if params.PageSize != nil {
			query.Set("page_size", strconv.Itoa(*params.PageSize))
		}
		if encoded := query.Encode(); encoded != "" {
			path += "?" + encoded
		}
	}

	var out ListChatRoomsResponse
	if err := client.Do(ctx, http.MethodGet, path, nil, &out); err != nil {
		return nil, err
	}

	return &out, nil
}
