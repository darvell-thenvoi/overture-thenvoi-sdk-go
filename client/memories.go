package client

import (
	"context"
	"errors"
	"net/http"
	"net/url"
)

// ListMemoriesInput contains filters for ListMemories.
type ListMemoriesInput struct {
	SubjectID    *string
	Scope        *string
	System       *string
	Type         *string
	Segment      *string
	ContentQuery *string
	PageSize     *int
	Status       *string
}

// ListMemoriesResponse contains memories and pagination metadata.
type ListMemoriesResponse struct {
	Data []Memory           `json:"data"`
	Meta PaginationMetadata `json:"meta"`
}

// CreateMemoryInput contains fields for storing a memory.
type CreateMemoryInput struct {
	Content   string         `json:"content"`
	Metadata  map[string]any `json:"metadata,omitempty"`
	Scope     *string        `json:"scope,omitempty"`
	Segment   string         `json:"segment"`
	SubjectID *string        `json:"subject_id,omitempty"`
	System    string         `json:"system"`
	Thought   string         `json:"thought"`
	Type      string         `json:"type"`
}

// ListMemories lists memories accessible to the agent.
func (client *Client) ListMemories(ctx context.Context, input *ListMemoriesInput) (*ListMemoriesResponse, error) {
	values := url.Values{}
	if input != nil {
		if input.SubjectID != nil {
			values.Set("subject_id", *input.SubjectID)
		}
		if input.Scope != nil {
			values.Set("scope", *input.Scope)
		}
		if input.System != nil {
			values.Set("system", *input.System)
		}
		if input.Type != nil {
			values.Set("type", *input.Type)
		}
		if input.Segment != nil {
			values.Set("segment", *input.Segment)
		}
		if input.ContentQuery != nil {
			values.Set("content_query", *input.ContentQuery)
		}
		if input.PageSize != nil {
			values.Set("page_size", encodeInt(*input.PageSize))
		}
		if input.Status != nil {
			values.Set("status", *input.Status)
		}
	}
	var out ListMemoriesResponse
	if err := client.Do(ctx, http.MethodGet, appendQuery("/api/v1/agent/memories", values), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// CreateMemory stores a memory entry.
func (client *Client) CreateMemory(ctx context.Context, input CreateMemoryInput) (*Memory, error) {
	if input.Content == "" {
		return nil, errors.New("thenvoi: content is required")
	}
	if input.Segment == "" {
		return nil, errors.New("thenvoi: segment is required")
	}
	if input.System == "" {
		return nil, errors.New("thenvoi: system is required")
	}
	if input.Thought == "" {
		return nil, errors.New("thenvoi: thought is required")
	}
	if input.Type == "" {
		return nil, errors.New("thenvoi: type is required")
	}
	var out struct {
		Data Memory `json:"data"`
	}
	if err := client.Do(ctx, http.MethodPost, "/api/v1/agent/memories", map[string]any{"memory": input}, &out); err != nil {
		return nil, err
	}
	return &out.Data, nil
}

// GetMemory fetches a memory by id.
func (client *Client) GetMemory(ctx context.Context, memoryID string) (*Memory, error) {
	if memoryID == "" {
		return nil, errors.New("thenvoi: memory id is required")
	}
	var out struct {
		Data Memory `json:"data"`
	}
	if err := client.Do(ctx, http.MethodGet, "/api/v1/agent/memories/"+url.PathEscape(memoryID), nil, &out); err != nil {
		return nil, err
	}
	return &out.Data, nil
}

// SupersedeMemory marks a memory as superseded.
func (client *Client) SupersedeMemory(ctx context.Context, memoryID string) (*Memory, error) {
	return client.memoryOperation(ctx, memoryID, "supersede")
}

// ArchiveMemory marks a memory as archived.
func (client *Client) ArchiveMemory(ctx context.Context, memoryID string) (*Memory, error) {
	return client.memoryOperation(ctx, memoryID, "archive")
}

func (client *Client) memoryOperation(ctx context.Context, memoryID string, operation string) (*Memory, error) {
	if memoryID == "" {
		return nil, errors.New("thenvoi: memory id is required")
	}
	var out struct {
		Data Memory `json:"data"`
	}
	path := "/api/v1/agent/memories/" + url.PathEscape(memoryID) + "/" + operation
	if err := client.Do(ctx, http.MethodPost, path, nil, &out); err != nil {
		return nil, err
	}
	return &out.Data, nil
}
