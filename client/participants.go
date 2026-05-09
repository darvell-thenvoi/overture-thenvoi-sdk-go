package client

import (
	"context"
	"errors"
	"net/http"
	"net/url"
)

// ParticipantRole is the participant's access level in a chat.
type ParticipantRole string

const (
	ParticipantRoleOwner  ParticipantRole = "owner"
	ParticipantRoleAdmin  ParticipantRole = "admin"
	ParticipantRoleMember ParticipantRole = "member"
)

// ChatParticipantType is the participant kind.
type ChatParticipantType string

const (
	ChatParticipantTypeUser  ChatParticipantType = "User"
	ChatParticipantTypeAgent ChatParticipantType = "Agent"
)

// ChatParticipant describes a chat room participant.
type ChatParticipant struct {
	Handle *string             `json:"handle"`
	ID     string              `json:"id"`
	Name   *string             `json:"name"`
	Role   ParticipantRole     `json:"role"`
	Status string              `json:"status"`
	Type   ChatParticipantType `json:"type"`
}

// AddChatParticipantInput contains fields for AddChatParticipant.
type AddChatParticipantInput struct {
	ParticipantID string          `json:"participant_id"`
	Role          ParticipantRole `json:"role,omitempty"`
}

// ListChatParticipants lists participants in an agent chat.
func (client *Client) ListChatParticipants(ctx context.Context, chatID string) ([]ChatParticipant, error) {
	if chatID == "" {
		return nil, errors.New("thenvoi: chat id is required")
	}

	escapedChatID := url.PathEscape(chatID)

	var out struct {
		Data []ChatParticipant `json:"data"`
	}

	err := client.Do(
		ctx,
		http.MethodGet,
		"/api/v1/agent/chats/"+escapedChatID+"/participants",
		nil,
		&out,
	)
	if err != nil {
		return nil, err
	}

	return out.Data, nil
}

// AddChatParticipant adds a participant to an agent chat.
func (client *Client) AddChatParticipant(ctx context.Context, chatID string, input AddChatParticipantInput) (*ChatParticipant, error) {
	if chatID == "" {
		return nil, errors.New("thenvoi: chat id is required")
	}
	if input.ParticipantID == "" {
		return nil, errors.New("thenvoi: participant id is required")
	}

	escapedChatID := url.PathEscape(chatID)

	var out struct {
		Data ChatParticipant `json:"data"`
	}

	err := client.Do(
		ctx,
		http.MethodPost,
		"/api/v1/agent/chats/"+escapedChatID+"/participants",
		map[string]any{"participant": input},
		&out,
	)
	if err != nil {
		return nil, err
	}

	return &out.Data, nil
}

// RemoveChatParticipant removes a participant from an agent chat.
func (client *Client) RemoveChatParticipant(ctx context.Context, chatID string, participantID string) (*ChatParticipant, error) {
	if chatID == "" {
		return nil, errors.New("thenvoi: chat id is required")
	}
	if participantID == "" {
		return nil, errors.New("thenvoi: participant id is required")
	}

	escapedChatID := url.PathEscape(chatID)
	escapedParticipantID := url.PathEscape(participantID)

	var out struct {
		Data ChatParticipant `json:"data"`
	}

	err := client.Do(
		ctx,
		http.MethodDelete,
		"/api/v1/agent/chats/"+escapedChatID+"/participants/"+escapedParticipantID,
		nil,
		&out,
	)
	if err != nil {
		return nil, err
	}

	return &out.Data, nil
}
