package client

import (
	"context"
	"errors"
	"net/http"
	"net/url"
)

// AddChatParticipantInput contains fields for adding a participant.
type AddChatParticipantInput struct {
	ParticipantID string  `json:"participant_id"`
	Role          *string `json:"role,omitempty"`
}

// ListChatParticipantsResponse contains chat participants.
type ListChatParticipantsResponse struct {
	Data []ChatParticipant `json:"data"`
}

// ListChatParticipants lists participants in a chat room.
func (client *Client) ListChatParticipants(ctx context.Context, chatID string) ([]ChatParticipant, error) {
	if chatID == "" {
		return nil, errors.New("band: chat id is required")
	}
	var out ListChatParticipantsResponse
	if err := client.Do(ctx, http.MethodGet, "/api/v1/agent/chats/"+url.PathEscape(chatID)+"/participants", nil, &out); err != nil {
		return nil, err
	}
	return out.Data, nil
}

// AddChatParticipant adds a participant to a chat room.
func (client *Client) AddChatParticipant(ctx context.Context, chatID string, input AddChatParticipantInput) (*ChatParticipant, error) {
	if chatID == "" {
		return nil, errors.New("band: chat id is required")
	}
	if input.ParticipantID == "" {
		return nil, errors.New("band: participant id is required")
	}
	var out struct {
		Data ChatParticipant `json:"data"`
	}
	body := map[string]any{"participant": input}
	if err := client.Do(ctx, http.MethodPost, "/api/v1/agent/chats/"+url.PathEscape(chatID)+"/participants", body, &out); err != nil {
		return nil, err
	}
	return &out.Data, nil
}

// RemoveChatParticipant removes a participant from a chat room.
func (client *Client) RemoveChatParticipant(ctx context.Context, chatID string, participantID string) (*ChatParticipant, error) {
	if chatID == "" {
		return nil, errors.New("band: chat id is required")
	}
	if participantID == "" {
		return nil, errors.New("band: participant id is required")
	}
	var out struct {
		Data ChatParticipant `json:"data"`
	}
	path := "/api/v1/agent/chats/" + url.PathEscape(chatID) + "/participants/" + url.PathEscape(participantID)
	if err := client.Do(ctx, http.MethodDelete, path, nil, &out); err != nil {
		return nil, err
	}
	return &out.Data, nil
}
