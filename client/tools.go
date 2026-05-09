package client

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"time"
)

// Tool describes a tool returned by the tool management API.
type Tool struct {
	ID               string           `json:"id"`
	Name             string           `json:"name"`
	Description      string           `json:"description"`
	JSONSchema       map[string]any   `json:"json_schema"`
	ConnectionConfig ConnectionConfig `json:"connection_config"`
	OwnerUUID        string           `json:"owner_uuid"`
	OrganizationID   *string          `json:"organization_id"`
	InsertedAt       time.Time        `json:"inserted_at"`
	UpdatedAt        time.Time        `json:"updated_at"`
}

// ToolListItem describes a tool summary returned by ListTools.
type ToolListItem struct {
	ID             string    `json:"id"`
	Name           string    `json:"name"`
	Description    string    `json:"description"`
	OwnerUUID      string    `json:"owner_uuid"`
	OrganizationID *string   `json:"organization_id"`
	InsertedAt     time.Time `json:"inserted_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// AssignedTool describes a tool assigned to another agent.
type AssignedTool struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	AssignedAt  time.Time `json:"assigned_at"`
}

// AssignedToolDetail describes a tool assigned to the current agent.
type AssignedToolDetail struct {
	ID          string         `json:"id"`
	Name        string         `json:"name"`
	Description string         `json:"description"`
	JSONSchema  map[string]any `json:"json_schema"`
	AssignedAt  time.Time      `json:"assigned_at"`
}

// ConnectionConfig describes how Thenvoi connects to a tool endpoint.
type ConnectionConfig struct {
	BaseURL   string     `json:"base_url"`
	Method    string     `json:"method"`
	Path      string     `json:"path"`
	ParamType string     `json:"param_type"`
	Auth      AuthConfig `json:"auth"`
}

// AuthConfig describes a tool endpoint authentication scheme.
type AuthConfig struct {
	Type       string  `json:"type"`
	Location   *string `json:"location,omitempty"`
	HeaderName *string `json:"header_name,omitempty"`
	KeyName    *string `json:"key_name,omitempty"`
}

// ToolPagination describes pagination metadata returned by tool endpoints.
type ToolPagination struct {
	Page       int `json:"page"`
	PerPage    int `json:"per_page"`
	TotalPages int `json:"total_pages"`
	TotalItems int `json:"total_items"`
}

// ListToolsInput contains pagination fields for ListTools.
type ListToolsInput struct {
	Page    *int
	PerPage *int
}

// ListToolsResponse contains tools and pagination metadata.
type ListToolsResponse struct {
	Data       []ToolListItem `json:"data"`
	Pagination ToolPagination `json:"pagination"`
}

// ListAgentToolsInput contains pagination fields for ListAgentTools.
type ListAgentToolsInput struct {
	Page    *int
	PerPage *int
}

// ListAgentToolsResponse contains assigned tools and pagination metadata.
type ListAgentToolsResponse struct {
	Data       []AssignedTool `json:"data"`
	Pagination ToolPagination `json:"pagination"`
}

// ListMyToolsInput contains pagination fields for ListMyTools.
type ListMyToolsInput struct {
	Page    *int
	PerPage *int
}

// ListMyToolsResponse contains tools assigned to the current agent.
type ListMyToolsResponse struct {
	Data       []AssignedToolDetail `json:"data"`
	Pagination ToolPagination       `json:"pagination"`
}

// CreateToolInput contains fields for creating a tool.
type CreateToolInput struct {
	Name             string            `json:"name"`
	Description      string            `json:"description"`
	JSONSchema       map[string]any    `json:"json_schema"`
	ConnectionConfig *ConnectionConfig `json:"connection_config"`
}

// UpdateToolInput contains fields for updating a tool.
type UpdateToolInput struct {
	Name             *string           `json:"name,omitempty"`
	Description      *string           `json:"description,omitempty"`
	JSONSchema       map[string]any    `json:"json_schema,omitempty"`
	ConnectionConfig *ConnectionConfig `json:"connection_config,omitempty"`
}

// AssignToolsInput contains tool ids to assign.
type AssignToolsInput struct {
	ToolIDs []string `json:"tool_ids"`
}

// AssignedToolSummary describes a tool assignment result item.
type AssignedToolSummary struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// AssignToolsToAgentResponse contains the tools assigned to an agent.
type AssignToolsToAgentResponse struct {
	AgentID       string                `json:"agent_id"`
	AssignedTools []AssignedToolSummary `json:"assigned_tools"`
}

// AssignToolsToMyselfResponse contains the tools assigned to the current agent.
type AssignToolsToMyselfResponse struct {
	AssignedTools []AssignedToolSummary `json:"assigned_tools"`
}

// ListTools lists tools visible to the authenticated caller.
func (client *Client) ListTools(ctx context.Context, input *ListToolsInput) (*ListToolsResponse, error) {
	values := url.Values{}
	if input != nil {
		addToolPagination(values, input.Page, input.PerPage)
	}
	var out ListToolsResponse
	if err := client.Do(ctx, http.MethodGet, appendQuery("/api/v2/tools", values), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// CreateTool creates a tool.
func (client *Client) CreateTool(ctx context.Context, input CreateToolInput) (*Tool, error) {
	if err := validateCreateToolInput(input); err != nil {
		return nil, err
	}
	var out struct {
		Data Tool `json:"data"`
	}
	if err := client.Do(ctx, http.MethodPost, "/api/v2/tools", input, &out); err != nil {
		return nil, err
	}
	return &out.Data, nil
}

// GetTool fetches a tool by id.
func (client *Client) GetTool(ctx context.Context, toolID string) (*Tool, error) {
	if toolID == "" {
		return nil, errors.New("thenvoi: tool id is required")
	}
	var out struct {
		Data Tool `json:"data"`
	}
	if err := client.Do(ctx, http.MethodGet, "/api/v2/tools/"+url.PathEscape(toolID), nil, &out); err != nil {
		return nil, err
	}
	return &out.Data, nil
}

// UpdateTool updates a tool by id.
func (client *Client) UpdateTool(ctx context.Context, toolID string, input UpdateToolInput) (*Tool, error) {
	if toolID == "" {
		return nil, errors.New("thenvoi: tool id is required")
	}
	if input.ConnectionConfig != nil {
		if err := validateConnectionConfig(*input.ConnectionConfig); err != nil {
			return nil, err
		}
	}
	var out struct {
		Data Tool `json:"data"`
	}
	if err := client.Do(ctx, http.MethodPut, "/api/v2/tools/"+url.PathEscape(toolID), input, &out); err != nil {
		return nil, err
	}
	return &out.Data, nil
}

// DeleteTool deletes a tool by id.
func (client *Client) DeleteTool(ctx context.Context, toolID string) error {
	if toolID == "" {
		return errors.New("thenvoi: tool id is required")
	}
	return client.Do(ctx, http.MethodDelete, "/api/v2/tools/"+url.PathEscape(toolID), nil, nil)
}

// AssignToolsToAgent assigns tools to an agent.
func (client *Client) AssignToolsToAgent(ctx context.Context, agentID string, input AssignToolsInput) (*AssignToolsToAgentResponse, error) {
	if agentID == "" {
		return nil, errors.New("thenvoi: agent id is required")
	}
	if err := validateAssignToolsInput(input); err != nil {
		return nil, err
	}
	var out struct {
		Data AssignToolsToAgentResponse `json:"data"`
	}
	path := "/api/v2/agents/" + url.PathEscape(agentID) + "/tools"
	if err := client.Do(ctx, http.MethodPost, path, input, &out); err != nil {
		return nil, err
	}
	return &out.Data, nil
}

// ListAgentTools lists tools assigned to an agent.
func (client *Client) ListAgentTools(ctx context.Context, agentID string, input *ListAgentToolsInput) (*ListAgentToolsResponse, error) {
	if agentID == "" {
		return nil, errors.New("thenvoi: agent id is required")
	}
	values := url.Values{}
	if input != nil {
		addToolPagination(values, input.Page, input.PerPage)
	}
	path := "/api/v2/agents/" + url.PathEscape(agentID) + "/tools"
	var out ListAgentToolsResponse
	if err := client.Do(ctx, http.MethodGet, appendQuery(path, values), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// RemoveToolFromAgent removes a tool assignment from an agent.
func (client *Client) RemoveToolFromAgent(ctx context.Context, agentID string, toolID string) error {
	if agentID == "" {
		return errors.New("thenvoi: agent id is required")
	}
	if toolID == "" {
		return errors.New("thenvoi: tool id is required")
	}
	path := "/api/v2/agents/" + url.PathEscape(agentID) + "/tools/" + url.PathEscape(toolID)
	return client.Do(ctx, http.MethodDelete, path, nil, nil)
}

// ListMyTools lists tools assigned to the current agent.
func (client *Client) ListMyTools(ctx context.Context, input *ListMyToolsInput) (*ListMyToolsResponse, error) {
	values := url.Values{}
	if input != nil {
		addToolPagination(values, input.Page, input.PerPage)
	}
	var out ListMyToolsResponse
	if err := client.Do(ctx, http.MethodGet, appendQuery("/api/v2/agents/me/tools", values), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// AssignToolsToMyself assigns tools to the current agent.
func (client *Client) AssignToolsToMyself(ctx context.Context, input AssignToolsInput) (*AssignToolsToMyselfResponse, error) {
	if err := validateAssignToolsInput(input); err != nil {
		return nil, err
	}
	var out struct {
		Data AssignToolsToMyselfResponse `json:"data"`
	}
	if err := client.Do(ctx, http.MethodPost, "/api/v2/agents/me/tools", input, &out); err != nil {
		return nil, err
	}
	return &out.Data, nil
}

// RemoveToolFromMyself removes a tool assignment from the current agent.
func (client *Client) RemoveToolFromMyself(ctx context.Context, toolID string) error {
	if toolID == "" {
		return errors.New("thenvoi: tool id is required")
	}
	return client.Do(ctx, http.MethodDelete, "/api/v2/agents/me/tools/"+url.PathEscape(toolID), nil, nil)
}

func addToolPagination(values url.Values, page *int, perPage *int) {
	if page != nil {
		values.Set("page", encodeInt(*page))
	}
	if perPage != nil {
		values.Set("per_page", encodeInt(*perPage))
	}
}

func validateCreateToolInput(input CreateToolInput) error {
	if input.Name == "" {
		return errors.New("thenvoi: tool name is required")
	}
	if input.Description == "" {
		return errors.New("thenvoi: tool description is required")
	}
	if input.JSONSchema == nil {
		return errors.New("thenvoi: tool json schema is required")
	}
	if input.ConnectionConfig == nil {
		return errors.New("thenvoi: tool connection config is required")
	}
	return validateConnectionConfig(*input.ConnectionConfig)
}

func validateConnectionConfig(config ConnectionConfig) error {
	if config.BaseURL == "" {
		return errors.New("thenvoi: tool connection base url is required")
	}
	if config.Method == "" {
		return errors.New("thenvoi: tool connection method is required")
	}
	if config.Path == "" {
		return errors.New("thenvoi: tool connection path is required")
	}
	if config.ParamType == "" {
		return errors.New("thenvoi: tool connection param type is required")
	}
	return validateAuthConfig(config.Auth)
}

func validateAuthConfig(config AuthConfig) error {
	switch config.Type {
	case "":
		return errors.New("thenvoi: tool auth type is required")
	case "api_key":
		if config.Location == nil || *config.Location == "" {
			return errors.New("thenvoi: tool auth location is required")
		}
		if config.HeaderName == nil || *config.HeaderName == "" {
			return errors.New("thenvoi: tool auth header name is required")
		}
		if config.KeyName == nil || *config.KeyName == "" {
			return errors.New("thenvoi: tool auth key name is required")
		}
	case "bearer", "basic", "vercel_bypass":
		if config.KeyName == nil || *config.KeyName == "" {
			return errors.New("thenvoi: tool auth key name is required")
		}
	}
	return nil
}

func validateAssignToolsInput(input AssignToolsInput) error {
	if len(input.ToolIDs) == 0 {
		return errors.New("thenvoi: tool ids are required")
	}
	for _, toolID := range input.ToolIDs {
		if toolID == "" {
			return errors.New("thenvoi: tool id is required")
		}
	}
	return nil
}
