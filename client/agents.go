package client

import (
	"context"
	"net/http"
)

// GetAgent fetches the current authenticated agent identity.
func (client *Client) GetAgent(ctx context.Context) (*AgentIdentity, error) {
	var out struct {
		Data AgentIdentity `json:"data"`
	}
	if err := client.Do(ctx, http.MethodGet, "/api/v1/agent/me", nil, &out); err != nil {
		return nil, err
	}
	return &out.Data, nil
}

// GetAgentMe fetches the current authenticated agent identity.
func (client *Client) GetAgentMe(ctx context.Context) (*AgentIdentity, error) {
	return client.GetAgent(ctx)
}
