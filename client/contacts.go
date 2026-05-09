package client

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"time"
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
	Data     ContactRequestsData      `json:"data"`
	Metadata *ContactRequestsMetadata `json:"metadata,omitempty"`
}

// ContactRequestsData contains received and sent contact requests.
type ContactRequestsData struct {
	Received []ContactRequest `json:"received,omitempty"`
	Sent     []ContactRequest `json:"sent,omitempty"`
}

// ContactRequestsMetadata contains per-direction contact request pagination metadata.
type ContactRequestsMetadata struct {
	Page     int                             `json:"page"`
	PageSize int                             `json:"page_size"`
	Received ContactRequestDirectionMetadata `json:"received"`
	Sent     ContactRequestDirectionMetadata `json:"sent"`
}

// ContactRequestDirectionMetadata contains pagination totals for one contact request direction.
type ContactRequestDirectionMetadata struct {
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}

// RespondContactRequestInput contains fields for approving, rejecting, or canceling a request.
type RespondContactRequestInput struct {
	Action    string  `json:"action"`
	Handle    *string `json:"handle,omitempty"`
	RequestID *string `json:"request_id,omitempty"`
}

// ContactOperationResponse describes contact mutation results.
type ContactOperationResponse struct {
	Data *ContactOperationResult `json:"data,omitempty"`
}

// ContactOperationResult is returned by contact mutation endpoints.
type ContactOperationResult struct {
	ID     string `json:"id,omitempty"`
	Status string `json:"status"`
}

// ContactSearchFilters contains filters for SearchContacts.
type ContactSearchFilters struct {
	Types        []string                 `json:"types,omitempty"`
	Capabilities *ContactCapabilityFilter `json:"capabilities,omitempty"`
	Tags         []string                 `json:"tags,omitempty"`
}

// ContactCapabilityFilter contains required or alternative contact capabilities.
type ContactCapabilityFilter struct {
	Required []string `json:"required,omitempty"`
	AnyOf    []string `json:"any_of,omitempty"`
}

// SearchContactsInput contains fields for searching API v2 contacts.
type SearchContactsInput struct {
	Query   string                `json:"query,omitempty"`
	Filters *ContactSearchFilters `json:"filters,omitempty"`
	Page    *int                  `json:"page,omitempty"`
	PerPage *int                  `json:"per_page,omitempty"`
}

// SearchContactsResponse contains contact search results and pagination.
type SearchContactsResponse struct {
	Data       []Contact  `json:"data"`
	Pagination Pagination `json:"pagination"`
}

// CheckContactAvailabilityInput contains contact availability request fields.
type CheckContactAvailabilityInput struct {
	ContactIDs              []string `json:"contact_ids"`
	Purpose                 string   `json:"purpose,omitempty"`
	RequiredDurationMinutes *int     `json:"required_duration_minutes,omitempty"`
}

// CheckContactAvailabilityResponse contains contact availability results.
type CheckContactAvailabilityResponse struct {
	Contacts  []ContactAvailability      `json:"contacts"`
	Summary   ContactAvailabilitySummary `json:"summary"`
	Timestamp time.Time                  `json:"timestamp"`
}

// ContactAvailabilitySummary describes aggregate availability counts.
type ContactAvailabilitySummary struct {
	TotalChecked  int `json:"total_checked"`
	AvailableNow  int `json:"available_now"`
	AvailableSoon int `json:"available_soon"`
	Unavailable   int `json:"unavailable"`
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
	return out.Data, nil
}

// RemoveContact removes a contact by handle or contact id.
func (client *Client) RemoveContact(ctx context.Context, input RemoveContactInput) (*ContactOperationResult, error) {
	if (input.Handle == nil || *input.Handle == "") && (input.ContactID == nil || *input.ContactID == "") {
		return nil, errors.New("thenvoi: handle or contact id is required")
	}
	var out ContactOperationResponse
	if err := client.Do(ctx, http.MethodPost, "/api/v1/agent/contacts/remove", input, &out); err != nil {
		return nil, err
	}
	return out.Data, nil
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
	if (input.Handle == nil || *input.Handle == "") && (input.RequestID == nil || *input.RequestID == "") {
		return nil, errors.New("thenvoi: handle or request id is required")
	}
	var out ContactOperationResponse
	if err := client.Do(ctx, http.MethodPost, "/api/v1/agent/contacts/requests/respond", input, &out); err != nil {
		return nil, err
	}
	return out.Data, nil
}

// GetContactDetails fetches API v2 contact details.
func (client *Client) GetContactDetails(ctx context.Context, contactID string) (*ContactDetail, error) {
	if contactID == "" {
		return nil, errors.New("thenvoi: contact id is required")
	}

	var out struct {
		Data ContactDetail `json:"data"`
	}
	path := "/api/v2/agents/me/contacts/" + url.PathEscape(contactID)
	if err := client.Do(ctx, http.MethodGet, path, nil, &out); err != nil {
		return nil, err
	}
	return &out.Data, nil
}

// SearchContacts searches API v2 contacts.
func (client *Client) SearchContacts(ctx context.Context, input SearchContactsInput) (*SearchContactsResponse, error) {
	var out SearchContactsResponse
	if err := client.Do(ctx, http.MethodPost, "/api/v2/agents/me/contacts/search", input, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// CheckContactAvailability checks availability for up to 50 API v2 contacts.
func (client *Client) CheckContactAvailability(ctx context.Context, input CheckContactAvailabilityInput) (*CheckContactAvailabilityResponse, error) {
	if len(input.ContactIDs) == 0 {
		return nil, errors.New("thenvoi: at least one contact id is required")
	}
	if len(input.ContactIDs) > 50 {
		return nil, errors.New("thenvoi: at most 50 contact ids are allowed")
	}

	var out struct {
		Data CheckContactAvailabilityResponse `json:"data"`
	}
	if err := client.Do(ctx, http.MethodPost, "/api/v2/agents/me/contacts/availability", input, &out); err != nil {
		return nil, err
	}
	return &out.Data, nil
}
