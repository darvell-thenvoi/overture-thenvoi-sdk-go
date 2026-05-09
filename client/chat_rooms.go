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

// ListChatRoomsOptions controls pagination for chat room listing.
type ListChatRoomsOptions struct {
	Page     *int
	PageSize *int
}

// ListChatRooms fetches chat rooms available to the authenticated agent.
func (client *Client) ListChatRooms(ctx context.Context) (*ListChatRoomsResponse, error) {
	return client.ListChatRoomsWithOptions(ctx, nil)
}

// ListChatRoomsWithOptions fetches chat rooms with optional pagination params.
func (client *Client) ListChatRoomsWithOptions(ctx context.Context, opts *ListChatRoomsOptions) (*ListChatRoomsResponse, error) {
	path := "/api/v1/agent/chats"
	if opts != nil {
		query := url.Values{}
		if opts.Page != nil {
			query.Set("page", strconv.Itoa(*opts.Page))
		}
		if opts.PageSize != nil {
			query.Set("page_size", strconv.Itoa(*opts.PageSize))
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
