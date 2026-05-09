package client

import (
	"context"
	"errors"
	"net/http"
	"net/url"
)

// ListChatRoomsInput contains filters for ListChatRooms.
type ListChatRoomsInput struct {
	Page     *int
	PageSize *int
}

// ListChatRoomsResponse contains chat rooms and pagination metadata.
type ListChatRoomsResponse struct {
	Data     []ChatRoom         `json:"data"`
	Metadata PaginationMetadata `json:"metadata"`
}

// CreateChatRoomInput contains fields for creating an agent-owned chat room.
type CreateChatRoomInput struct {
	TaskID *string `json:"task_id,omitempty"`
}

// ListChatRooms lists chat rooms where the current agent is a participant.
func (client *Client) ListChatRooms(ctx context.Context, input *ListChatRoomsInput) (*ListChatRoomsResponse, error) {
	values := url.Values{}
	if input != nil {
		addPagination(values, PageInput{Page: input.Page, PageSize: input.PageSize})
	}

	var out ListChatRoomsResponse
	if err := client.Do(ctx, http.MethodGet, appendQuery("/api/v1/agent/chats", values), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetChatRoom fetches a chat room where the current agent is a participant.
func (client *Client) GetChatRoom(ctx context.Context, chatID string) (*ChatRoom, error) {
	if chatID == "" {
		return nil, errors.New("thenvoi: chat id is required")
	}

	var out struct {
		Data ChatRoom `json:"data"`
	}
	if err := client.Do(ctx, http.MethodGet, "/api/v1/agent/chats/"+url.PathEscape(chatID), nil, &out); err != nil {
		return nil, err
	}
	return &out.Data, nil
}

// CreateChatRoom creates a chat room with the current agent as owner.
func (client *Client) CreateChatRoom(ctx context.Context, input CreateChatRoomInput) (*ChatRoom, error) {
	var out struct {
		Data ChatRoom `json:"data"`
	}
	if err := client.Do(ctx, http.MethodPost, "/api/v1/agent/chats", map[string]any{"chat": input}, &out); err != nil {
		return nil, err
	}
	return &out.Data, nil
}
