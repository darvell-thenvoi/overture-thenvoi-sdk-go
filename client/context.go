package client

import (
	"context"
	"errors"
	"net/http"
	"net/url"
)

// GetChatContextInput contains filters for GetChatContext.
type GetChatContextInput struct {
	Page     *int
	PageSize *int
}

// GetChatContextResponse contains context messages and pagination metadata.
type GetChatContextResponse struct {
	Data []ChatMessage      `json:"data"`
	Meta PaginationMetadata `json:"meta"`
}

// GetChatContext returns messages relevant to the agent for execution context.
func (client *Client) GetChatContext(ctx context.Context, chatID string, input *GetChatContextInput) (*GetChatContextResponse, error) {
	if chatID == "" {
		return nil, errors.New("thenvoi: chat id is required")
	}
	values := url.Values{}
	if input != nil {
		addPagination(values, PageInput{Page: input.Page, PageSize: input.PageSize})
	}
	path := "/api/v1/agent/chats/" + url.PathEscape(chatID) + "/context"
	var out GetChatContextResponse
	if err := client.Do(ctx, http.MethodGet, appendQuery(path, values), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
