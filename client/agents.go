package client

import (
	"context"
	"net/http"
)

// AgentIdentity describes the authenticated agent profile.
type AgentIdentity struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Description *string `json:"description"`
}

// GetAgent fetches the current authenticated agent identity.
func (client *Client) GetAgent(ctx context.Context) (*AgentIdentity, error) {
	var out AgentIdentity
	if err := client.Do(ctx, http.MethodGet, "/api/v1/agent/me", nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetAgentMe fetches the current authenticated agent identity.
func (client *Client) GetAgentMe(ctx context.Context) (*AgentIdentity, error) {
	return client.GetAgent(ctx)
}
