package client

import (
	"context"
	"errors"
	"net/http"
	"net/url"
)

// ListContactsInput contains filters for ListContacts.
type ListContactsInput struct {
	Page     *int
	PageSize *int
}

// ListContactsResponse contains contacts and pagination metadata.
type ListContactsResponse struct {
	Data     []AgentContact     `json:"data"`
	Metadata PaginationMetadata `json:"metadata"`
}

// AddContactInput contains fields for sending a contact request.
type AddContactInput struct {
	Handle  string  `json:"handle"`
	Message *string `json:"message,omitempty"`
}

// RemoveContactInput removes a contact by handle or contact id.
type RemoveContactInput struct {
	Handle    *string `json:"handle,omitempty"`
	ContactID *string `json:"contact_id,omitempty"`
}

// ListContactRequestsInput contains filters for contact request listing.
type ListContactRequestsInput struct {
	Page       *int
	PageSize   *int
	SentStatus *string
}

// ListContactRequestsResponse contains received and sent contact requests.
type ListContactRequestsResponse struct {
	Data     ContactRequestsData `json:"data"`
	Metadata map[string]any      `json:"metadata,omitempty"`
}

// ContactRequestsData contains received and sent contact requests.
type ContactRequestsData struct {
	Received []ContactRequest `json:"received,omitempty"`
	Sent     []ContactRequest `json:"sent,omitempty"`
}

// RespondContactRequestInput contains fields for approving, rejecting, or canceling a request.
type RespondContactRequestInput struct {
	Action    string  `json:"action"`
	Handle    *string `json:"handle,omitempty"`
	RequestID *string `json:"request_id,omitempty"`
}

// ContactOperationResponse describes contact mutation results.
type ContactOperationResponse struct {
	Data ContactOperationResult `json:"data,omitempty"`
}

// ContactOperationResult is returned by contact mutation endpoints.
type ContactOperationResult struct {
	ID     string `json:"id,omitempty"`
	Status string `json:"status"`
}

// ListContacts lists the current agent's contacts.
func (client *Client) ListContacts(ctx context.Context, input *ListContactsInput) (*ListContactsResponse, error) {
	values := url.Values{}
	if input != nil {
		addPagination(values, PageInput{Page: input.Page, PageSize: input.PageSize})
	}
	var out ListContactsResponse
	if err := client.Do(ctx, http.MethodGet, appendQuery("/api/v1/agent/contacts", values), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// AddContact sends a contact request by handle.
func (client *Client) AddContact(ctx context.Context, input AddContactInput) (*ContactOperationResult, error) {
	if input.Handle == "" {
		return nil, errors.New("thenvoi: handle is required")
	}
	var out ContactOperationResponse
	if err := client.Do(ctx, http.MethodPost, "/api/v1/agent/contacts/add", input, &out); err != nil {
		return nil, err
	}
	return &out.Data, nil
}

// RemoveContact removes a contact by handle or contact id.
func (client *Client) RemoveContact(ctx context.Context, input RemoveContactInput) (*ContactOperationResult, error) {
	if input.Handle == nil && input.ContactID == nil {
		return nil, errors.New("thenvoi: handle or contact id is required")
	}
	var out ContactOperationResponse
	if err := client.Do(ctx, http.MethodPost, "/api/v1/agent/contacts/remove", input, &out); err != nil {
		return nil, err
	}
	return &out.Data, nil
}

// ListContactRequests lists received and sent contact requests.
func (client *Client) ListContactRequests(ctx context.Context, input *ListContactRequestsInput) (*ListContactRequestsResponse, error) {
	values := url.Values{}
	if input != nil {
		addPagination(values, PageInput{Page: input.Page, PageSize: input.PageSize})
		if input.SentStatus != nil {
			values.Set("sent_status", *input.SentStatus)
		}
	}
	var out ListContactRequestsResponse
	if err := client.Do(ctx, http.MethodGet, appendQuery("/api/v1/agent/contacts/requests", values), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// RespondContactRequest approves, rejects, or cancels a contact request.
func (client *Client) RespondContactRequest(ctx context.Context, input RespondContactRequestInput) (*ContactOperationResult, error) {
	if input.Action == "" {
		return nil, errors.New("thenvoi: action is required")
	}
	var out ContactOperationResponse
	if err := client.Do(ctx, http.MethodPost, "/api/v1/agent/contacts/requests/respond", input, &out); err != nil {
		return nil, err
	}
	return &out.Data, nil
}
