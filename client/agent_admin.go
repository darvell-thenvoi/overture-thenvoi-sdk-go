package client

import (
	"context"
	"errors"
	"net/http"
	"net/url"
)

// ListAgentsInput contains pagination filters for ListAgents.
type ListAgentsInput struct {
	Page    *int
	PerPage *int
}

// ListAgentsResponse contains API v2 agents and pagination.
type ListAgentsResponse struct {
	Data       []AgentListItem `json:"data"`
	Pagination Pagination      `json:"pagination"`
}

// RegisterExternalAgentInput contains fields for registering an external agent.
type RegisterExternalAgentInput struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	ModelType   string `json:"model_type"`
}

// RegisterExternalAgentResponse contains the registered agent and credentials.
type RegisterExternalAgentResponse struct {
	Agent       Agent                       `json:"agent"`
	Credentials RegisterExternalCredentials `json:"credentials"`
}

// RegisterExternalCredentials contains credentials returned for external agents.
type RegisterExternalCredentials struct {
	APIKey  string `json:"api_key"`
	Message string `json:"message"`
}

// CreatePlatformAgentInput contains fields for creating a platform agent.
type CreatePlatformAgentInput struct {
	Name                   string         `json:"name"`
	Description            string         `json:"description"`
	ModelType              string         `json:"model_type"`
	SystemPrompt           string         `json:"system_prompt,omitempty"`
	ToolIDs                []string       `json:"tool_ids,omitempty"`
	StructuredOutputSchema map[string]any `json:"structured_output_schema,omitempty"`
}

// UpdateAgentInput contains fields for updating an agent.
type UpdateAgentInput struct {
	Name                   string         `json:"name,omitempty"`
	Description            string         `json:"description,omitempty"`
	ModelType              string         `json:"model_type,omitempty"`
	SystemPrompt           string         `json:"system_prompt,omitempty"`
	StructuredOutputSchema map[string]any `json:"structured_output_schema,omitempty"`
}

// ExecuteAgentInput contains fields for requesting an agent execution.
type ExecuteAgentInput struct {
	Request  string         `json:"request"`
	Summary  string         `json:"summary,omitempty"`
	Metadata map[string]any `json:"metadata,omitempty"`
}

// UpdateCurrentAgentInput contains fields for updating the authenticated agent.
type UpdateCurrentAgentInput struct {
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
}

// ListAgentChatRoomsInput contains pagination filters for ListAgentChatRooms.
type ListAgentChatRoomsInput struct {
	Page    *int
	PerPage *int
}

// ListAgentChatRoomsResponse contains current-agent chat rooms and pagination.
type ListAgentChatRoomsResponse struct {
	Data       []ChatRoom `json:"data"`
	Pagination Pagination `json:"pagination"`
}

// ListAgents lists API v2 agents.
func (client *Client) ListAgents(ctx context.Context, input *ListAgentsInput) (*ListAgentsResponse, error) {
	values := url.Values{}
	if input != nil {
		addV2Pagination(values, V2PageInput{Page: input.Page, PerPage: input.PerPage})
	}

	var out ListAgentsResponse
	if err := client.Do(ctx, http.MethodGet, appendQuery("/api/v2/agents", values), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// RegisterExternalAgent registers an external API v2 agent.
func (client *Client) RegisterExternalAgent(ctx context.Context, input RegisterExternalAgentInput) (*RegisterExternalAgentResponse, error) {
	if input.Name == "" {
		return nil, errors.New("thenvoi: name is required")
	}
	if input.Description == "" {
		return nil, errors.New("thenvoi: description is required")
	}
	if input.ModelType == "" {
		return nil, errors.New("thenvoi: model type is required")
	}

	var out struct {
		Data RegisterExternalAgentResponse `json:"data"`
	}
	if err := client.Do(ctx, http.MethodPost, "/api/v2/agents/register", input, &out); err != nil {
		return nil, err
	}
	return &out.Data, nil
}

// CreatePlatformAgent creates an API v2 platform agent.
func (client *Client) CreatePlatformAgent(ctx context.Context, input CreatePlatformAgentInput) (*Agent, error) {
	if input.Name == "" {
		return nil, errors.New("thenvoi: name is required")
	}
	if input.Description == "" {
		return nil, errors.New("thenvoi: description is required")
	}
	if input.ModelType == "" {
		return nil, errors.New("thenvoi: model type is required")
	}

	var out struct {
		Data Agent `json:"data"`
	}
	if err := client.Do(ctx, http.MethodPost, "/api/v2/agents", input, &out); err != nil {
		return nil, err
	}
	return &out.Data, nil
}

// UpdateAgent updates an API v2 agent by id.
func (client *Client) UpdateAgent(ctx context.Context, agentID string, input UpdateAgentInput) (*Agent, error) {
	if agentID == "" {
		return nil, errors.New("thenvoi: agent id is required")
	}

	var out struct {
		Data Agent `json:"data"`
	}
	if err := client.Do(ctx, http.MethodPut, "/api/v2/agents/"+url.PathEscape(agentID), input, &out); err != nil {
		return nil, err
	}
	return &out.Data, nil
}

// DeleteAgent deletes an API v2 agent by id.
func (client *Client) DeleteAgent(ctx context.Context, agentID string) error {
	if agentID == "" {
		return errors.New("thenvoi: agent id is required")
	}
	return client.Do(ctx, http.MethodDelete, "/api/v2/agents/"+url.PathEscape(agentID), nil, nil)
}

// ExecuteAgent starts an API v2 agent execution.
func (client *Client) ExecuteAgent(ctx context.Context, agentID string, input ExecuteAgentInput) (*ExecutionResponse, error) {
	if agentID == "" {
		return nil, errors.New("thenvoi: agent id is required")
	}
	if input.Request == "" {
		return nil, errors.New("thenvoi: request is required")
	}

	var out struct {
		Data ExecutionResponse `json:"data"`
	}
	if err := client.Do(ctx, http.MethodPost, "/api/v2/agents/"+url.PathEscape(agentID)+"/execute", input, &out); err != nil {
		return nil, err
	}
	return &out.Data, nil
}

// UpdateCurrentAgent updates the authenticated API v2 agent.
func (client *Client) UpdateCurrentAgent(ctx context.Context, input UpdateCurrentAgentInput) (*Agent, error) {
	var out struct {
		Data Agent `json:"data"`
	}
	if err := client.Do(ctx, http.MethodPut, "/api/v2/agents/me", input, &out); err != nil {
		return nil, err
	}
	return &out.Data, nil
}

// ListAgentChatRooms lists chat rooms for the authenticated API v2 agent.
func (client *Client) ListAgentChatRooms(ctx context.Context, input *ListAgentChatRoomsInput) (*ListAgentChatRoomsResponse, error) {
	values := url.Values{}
	if input != nil {
		addV2Pagination(values, V2PageInput{Page: input.Page, PerPage: input.PerPage})
	}

	var out ListAgentChatRoomsResponse
	if err := client.Do(ctx, http.MethodGet, appendQuery("/api/v2/agents/me/chats", values), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
