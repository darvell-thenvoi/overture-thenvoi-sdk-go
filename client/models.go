package client

import (
	"context"
	"net/http"
	"time"
)

// ListModelsResponse contains API v2 models and model metadata.
type ListModelsResponse struct {
	Data     []Model            `json:"data"`
	Metadata ListModelsMetadata `json:"metadata"`
}

// ListModelsMetadata describes API v2 model list metadata.
type ListModelsMetadata struct {
	TotalModels int       `json:"total_models"`
	LastUpdated time.Time `json:"last_updated"`
}

// ListModels lists API v2 models.
func (client *Client) ListModels(ctx context.Context) (*ListModelsResponse, error) {
	var out ListModelsResponse
	if err := client.Do(ctx, http.MethodGet, "/api/v2/models", nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
