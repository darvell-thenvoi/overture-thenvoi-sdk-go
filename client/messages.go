package client

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"time"
)

// Mention describes a user mention in a chat message.
type Mention struct {
	ID       string `json:"id"`
	Handle   string `json:"handle,omitempty"`
	Name     string `json:"name,omitempty"`
	Username string `json:"username,omitempty"`
}

// ChatMessage describes a chat message returned by the platform.
type ChatMessage struct {
	ID          string         `json:"id"`
	Content     string         `json:"content"`
	SenderID    string         `json:"sender_id"`
	SenderType  string         `json:"sender_type"`
	SenderName  *string        `json:"sender_name"`
	MessageType string         `json:"message_type"`
	Metadata    map[string]any `json:"metadata,omitempty"`
	InsertedAt  time.Time      `json:"inserted_at"`
	UpdatedAt   *time.Time     `json:"updated_at"`
}

// SendChatMessageInput contains message fields for SendChatMessage.
type SendChatMessageInput struct {
	Content     string         `json:"content"`
	MessageType string         `json:"message_type,omitempty"`
	Metadata    map[string]any `json:"metadata,omitempty"`
	Mentions    []Mention      `json:"mentions,omitempty"`
}

// MessageRecipient describes a recipient in a sent message acknowledgement.
type MessageRecipient struct {
	ID     string  `json:"id"`
	Handle string  `json:"handle"`
	Name   *string `json:"name,omitempty"`
}

// MessageSentResponse describes the send acknowledgement returned by the API.
type MessageSentResponse struct {
	ID         string             `json:"id"`
	Success    bool               `json:"success"`
	Recipients []MessageRecipient `json:"recipients,omitempty"`
}

// SendChatMessage sends a message to a chat room.
func (client *Client) SendChatMessage(ctx context.Context, chatID string, input SendChatMessageInput) (*MessageSentResponse, error) {
	if chatID == "" {
		return nil, errors.New("thenvoi: chat id is required")
	}
	if input.Content == "" {
		return nil, errors.New("thenvoi: content is required")
	}

	escapedChatID := url.PathEscape(chatID)

	var out struct {
		Data MessageSentResponse `json:"data"`
	}

	err := client.Do(
		ctx,
		http.MethodPost,
		"/api/v1/agent/chats/"+escapedChatID+"/messages",
		map[string]any{"message": input},
		&out,
	)
	if err != nil {
		return nil, err
	}

	return &out.Data, nil
}
