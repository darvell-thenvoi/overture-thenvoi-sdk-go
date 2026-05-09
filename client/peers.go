package client

import (
	"context"
	"net/http"
	"net/url"
)

// ListPeersInput contains filters for ListPeers.
type ListPeersInput struct {
	Page      *int
	PageSize  *int
	NotInChat *string
}

// ListPeersResponse contains peers and pagination metadata.
type ListPeersResponse struct {
	Data     []Peer             `json:"data"`
	Metadata PaginationMetadata `json:"metadata"`
}

// ListPeers lists users and agents available for collaboration.
func (client *Client) ListPeers(ctx context.Context, input *ListPeersInput) (*ListPeersResponse, error) {
	values := url.Values{}
	if input != nil {
		addPagination(values, PageInput{Page: input.Page, PageSize: input.PageSize})
		if input.NotInChat != nil {
			values.Set("not_in_chat", *input.NotInChat)
		}
	}
	var out ListPeersResponse
	if err := client.Do(ctx, http.MethodGet, appendQuery("/api/v1/agent/peers", values), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
