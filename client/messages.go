package client

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"strconv"
)

// SendChatMessageInput contains fields for the text-message endpoint.
type SendChatMessageInput struct {
	Content  string    `json:"content"`
	Mentions []Mention `json:"mentions"`
}

// CreateChatEventInput contains fields for the event endpoint.
type CreateChatEventInput struct {
	Content     string         `json:"content"`
	MessageType string         `json:"message_type"`
	Metadata    map[string]any `json:"metadata,omitempty"`
}

// ListChatMessagesInput contains filters for ListChatMessages.
type ListChatMessagesInput struct {
	Page     *int
	PageSize *int
	Status   *string
}

// ListChatMessagesResponse contains messages and pagination metadata.
type ListChatMessagesResponse struct {
	Data     []ChatMessage      `json:"data"`
	Metadata PaginationMetadata `json:"metadata"`
}

// SendChatMessage sends a text message to a chat room.
func (client *Client) SendChatMessage(ctx context.Context, chatID string, input SendChatMessageInput) (*MessageSentResponse, error) {
	if chatID == "" {
		return nil, errors.New("thenvoi: chat id is required")
	}
	if input.Content == "" {
		return nil, errors.New("thenvoi: content is required")
	}
	if len(input.Mentions) == 0 {
		return nil, errors.New("thenvoi: at least one mention is required")
	}
	for _, mention := range input.Mentions {
		if mention.ID == "" {
			return nil, errors.New("thenvoi: mention id is required")
		}
	}

	var out struct {
		Data MessageSentResponse `json:"data"`
	}
	err := client.Do(
		ctx,
		http.MethodPost,
		"/api/v1/agent/chats/"+url.PathEscape(chatID)+"/messages",
		map[string]any{"message": input},
		&out,
	)
	if err != nil {
		return nil, err
	}
	return &out.Data, nil
}

// CreateChatEvent records a non-text event in a chat room.
func (client *Client) CreateChatEvent(ctx context.Context, chatID string, input CreateChatEventInput) (*EventCreatedResponse, error) {
	if chatID == "" {
		return nil, errors.New("thenvoi: chat id is required")
	}
	if input.Content == "" {
		return nil, errors.New("thenvoi: content is required")
	}
	if input.MessageType == "" {
		return nil, errors.New("thenvoi: message type is required")
	}

	var out struct {
		Data EventCreatedResponse `json:"data"`
	}
	err := client.Do(
		ctx,
		http.MethodPost,
		"/api/v1/agent/chats/"+url.PathEscape(chatID)+"/events",
		map[string]any{"event": input},
		&out,
	)
	if err != nil {
		return nil, err
	}
	return &out.Data, nil
}

// ListChatMessages lists messages that the agent can process.
func (client *Client) ListChatMessages(ctx context.Context, chatID string, input *ListChatMessagesInput) (*ListChatMessagesResponse, error) {
	if chatID == "" {
		return nil, errors.New("thenvoi: chat id is required")
	}
	values := url.Values{}
	if input != nil {
		addPagination(values, PageInput{Page: input.Page, PageSize: input.PageSize})
		if input.Status != nil {
			values.Set("status", *input.Status)
		}
	}

	var out ListChatMessagesResponse
	path := "/api/v1/agent/chats/" + url.PathEscape(chatID) + "/messages"
	if err := client.Do(ctx, http.MethodGet, appendQuery(path, values), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetNextChatMessage returns the oldest unprocessed message for the agent, or nil on 204.
func (client *Client) GetNextChatMessage(ctx context.Context, chatID string) (*ChatMessage, error) {
	if chatID == "" {
		return nil, errors.New("thenvoi: chat id is required")
	}
	var out struct {
		Data ChatMessage `json:"data"`
	}
	status, err := client.do(ctx, http.MethodGet, "/api/v1/agent/chats/"+url.PathEscape(chatID)+"/messages/next", nil, &out)
	if err != nil {
		return nil, err
	}
	if status == http.StatusNoContent {
		return nil, nil
	}
	return &out.Data, nil
}

// MarkChatMessageProcessing marks a message as processing.
func (client *Client) MarkChatMessageProcessing(ctx context.Context, chatID string, messageID string) (*MessageStatusResponse, error) {
	return client.markChatMessageStatus(ctx, chatID, messageID, "processing", nil)
}

// MarkChatMessageProcessed marks a message as processed.
func (client *Client) MarkChatMessageProcessed(ctx context.Context, chatID string, messageID string) (*MessageStatusResponse, error) {
	return client.markChatMessageStatus(ctx, chatID, messageID, "processed", nil)
}

// MarkChatMessageFailed marks a message processing attempt as failed.
func (client *Client) MarkChatMessageFailed(ctx context.Context, chatID string, messageID string, reason string) (*MessageStatusResponse, error) {
	if reason == "" {
		return nil, errors.New("thenvoi: failure reason is required")
	}
	return client.markChatMessageStatus(ctx, chatID, messageID, "failed", map[string]string{"error": reason})
}

// DeleteChatMessage deletes an API v2 chat message.
func (client *Client) DeleteChatMessage(ctx context.Context, chatID string, messageID string) error {
	if chatID == "" {
		return errors.New("thenvoi: chat id is required")
	}
	if messageID == "" {
		return errors.New("thenvoi: message id is required")
	}

	path := "/api/v2/chats/" + url.PathEscape(chatID) + "/messages/" + url.PathEscape(messageID)
	return client.Do(ctx, http.MethodDelete, path, nil, nil)
}

func (client *Client) markChatMessageStatus(ctx context.Context, chatID string, messageID string, status string, body any) (*MessageStatusResponse, error) {
	if chatID == "" {
		return nil, errors.New("thenvoi: chat id is required")
	}
	if messageID == "" {
		return nil, errors.New("thenvoi: message id is required")
	}
	var out struct {
		Data MessageStatusResponse `json:"data"`
	}
	path := "/api/v1/agent/chats/" + url.PathEscape(chatID) + "/messages/" + url.PathEscape(messageID) + "/" + status
	if err := client.Do(ctx, http.MethodPost, path, body, &out); err != nil {
		return nil, err
	}
	return &out.Data, nil
}

func encodeInt(value int) string {
	return strconv.Itoa(value)
}
