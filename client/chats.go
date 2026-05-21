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

// ListMyChatsInput contains filters for ListMyChats.
type ListMyChatsInput struct {
	Status  *string
	Page    *int
	PerPage *int
}

// ListMyChatsResponse contains API v2 chat rooms and pagination metadata.
type ListMyChatsResponse struct {
	Data     []ChatRoom           `json:"data"`
	Metadata V2PaginationMetadata `json:"metadata"`
}

// CreateChatRoomInput contains fields for creating an agent-owned chat room.
type CreateChatRoomInput struct {
	TaskID *string `json:"task_id,omitempty"`
}

// UpdateChatRoomInput contains fields for updating an API v2 chat room.
type UpdateChatRoomInput struct {
	Title    string         `json:"title,omitempty"`
	Status   string         `json:"status,omitempty"`
	Metadata map[string]any `json:"metadata,omitempty"`
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

// ListMyChats lists chat rooms visible to the current user.
func (client *Client) ListMyChats(ctx context.Context, input *ListMyChatsInput) (*ListMyChatsResponse, error) {
	values := url.Values{}
	if input != nil {
		addV2Pagination(values, V2PageInput{Page: input.Page, PerPage: input.PerPage})
		if input.Status != nil {
			values.Set("status", *input.Status)
		}
	}

	var out ListMyChatsResponse
	if err := client.Do(ctx, http.MethodGet, appendQuery("/api/v2/me/chats", values), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetChatRoom fetches a chat room where the current agent is a participant.
func (client *Client) GetChatRoom(ctx context.Context, chatID string) (*ChatRoom, error) {
	if chatID == "" {
		return nil, errors.New("band: chat id is required")
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

// UpdateChatRoom updates an API v2 chat room.
func (client *Client) UpdateChatRoom(ctx context.Context, chatID string, input UpdateChatRoomInput) (*ChatRoom, error) {
	if chatID == "" {
		return nil, errors.New("band: chat id is required")
	}

	var out struct {
		Data ChatRoom `json:"data"`
	}
	path := "/api/v2/chats/" + url.PathEscape(chatID)
	if err := client.Do(ctx, http.MethodPut, path, map[string]any{"chat": input}, &out); err != nil {
		return nil, err
	}
	return &out.Data, nil
}

// DeleteChatRoom deletes an API v2 chat room.
func (client *Client) DeleteChatRoom(ctx context.Context, chatID string) error {
	if chatID == "" {
		return errors.New("band: chat id is required")
	}
	return client.Do(ctx, http.MethodDelete, "/api/v2/chats/"+url.PathEscape(chatID), nil, nil)
}
