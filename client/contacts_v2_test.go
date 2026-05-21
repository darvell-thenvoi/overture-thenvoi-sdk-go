package client_test

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/darvell-thenvoi/overture-thenvoi-sdk-go/client"
)

func TestContactV2Endpoints(t *testing.T) {
	t.Parallel()
	page := 2
	perPage := 25
	duration := 30
	seen := []string{}
	transport := roundTripFunc(func(req *http.Request) (*http.Response, error) {
		seen = append(seen, req.Method+" "+req.URL.String())
		switch req.URL.EscapedPath() {
		case "/api/v2/agents/me/contacts/contact%2F1":
			return jsonResponse(http.StatusOK, `{"data":{"id":"contact/1","name":"Ada","type":"agent","description":"Analyst","email":null,"avatar_url":null,"can_add_to_chats":true,"can_execute_requests":true,"status":{"connection":"online","availability":"available","last_seen":"2026-01-02T03:04:05Z","status_message":"ready"},"workload":{"active":1},"capabilities":["research"],"metadata":{"tier":"gold"},"relationship":{"added_at":"2026-01-01T03:04:05Z","interaction_count":7,"last_interaction":"2026-01-02T03:04:05Z","tags":["trusted"],"permissions":{"can_add_to_chats":true,"can_remove_from_chats":true,"can_execute_requests":true,"can_view_performance":false}},"created_at":"2026-01-01T03:04:05Z","performance_stats":{"success_rate":0.98}}}`), nil
		case "/api/v2/agents/me/contacts/search":
			assertJSONField(t, req, "query", "ada")
			return jsonResponse(http.StatusOK, `{"data":[{"id":"contact_1","name":"Ada","type":"agent","description":null,"email":null,"avatar_url":null,"can_add_to_chats":true,"can_execute_requests":true,"status":{"connection":"online","last_seen":null},"workload":null,"capabilities":["research"],"metadata":null,"relationship":{"added_at":"2026-01-01T03:04:05Z","interaction_count":7}}],"pagination":{"page":2,"per_page":25,"total_pages":1,"total_items":1}}`), nil
		case "/api/v2/agents/me/contacts/availability":
			var payload struct {
				ContactIDs []string `json:"contact_ids"`
				Purpose    string   `json:"purpose"`
			}
			decodeJSONBody(t, req, &payload)
			if len(payload.ContactIDs) != 2 || payload.ContactIDs[0] != "contact_1" || payload.Purpose != "chat" {
				t.Fatalf("availability payload=%#v", payload)
			}
			return jsonResponse(http.StatusOK, `{"data":{"contacts":[{"id":"contact_1","available":true,"status":{"connection":"online","availability":"now"},"estimated_wait_time":null,"confidence":0.9,"unavailable_reason":null}],"summary":{"total_checked":2,"available_now":1,"available_soon":0,"unavailable":1},"timestamp":"2026-01-02T03:04:05Z"}}`), nil
		default:
			t.Fatalf("unexpected path %s", req.URL.String())
			return nil, nil
		}
	})
	sdk := client.New(client.WithApiKey("test-key"), client.WithBaseURL("https://api.test"), client.WithHTTPClient(&http.Client{Transport: transport}))

	detail, err := sdk.GetContactDetails(context.Background(), "contact/1")
	if err != nil || detail.ID != "contact/1" || detail.Status.StatusMessage == nil || *detail.Status.StatusMessage != "ready" || detail.Relationship.Permissions == nil || !detail.Relationship.Permissions.CanExecuteRequests || detail.Relationship.LastInteraction == nil || len(detail.Relationship.Tags) != 1 {
		t.Fatalf("GetContactDetails out=%#v err=%v", detail, err)
	}
	search, err := sdk.SearchContacts(context.Background(), client.SearchContactsInput{Query: "ada", Filters: &client.ContactSearchFilters{Types: []string{"agent"}, Capabilities: &client.ContactCapabilityFilter{Required: []string{"research"}}}, Page: &page, PerPage: &perPage})
	if err != nil || len(search.Data) != 1 || search.Pagination.TotalItems != 1 {
		t.Fatalf("SearchContacts out=%#v err=%v", search, err)
	}
	availability, err := sdk.CheckContactAvailability(context.Background(), client.CheckContactAvailabilityInput{ContactIDs: []string{"contact_1", "contact_2"}, Purpose: "chat", RequiredDurationMinutes: &duration})
	if err != nil || availability.Summary.AvailableNow != 1 || len(availability.Contacts) != 1 {
		t.Fatalf("CheckContactAvailability out=%#v err=%v", availability, err)
	}

	want := []string{
		"GET https://api.test/api/v2/agents/me/contacts/contact%2F1",
		"POST https://api.test/api/v2/agents/me/contacts/search",
		"POST https://api.test/api/v2/agents/me/contacts/availability",
	}
	if len(seen) != len(want) {
		t.Fatalf("seen=%#v", seen)
	}
	for i := range want {
		if seen[i] != want[i] {
			t.Fatalf("seen[%d]=%s want %s", i, seen[i], want[i])
		}
	}
}

func TestContactV2ValidationAndAPIErrors(t *testing.T) {
	t.Parallel()
	sdk := client.New(client.WithApiKey("test-key"), client.WithBaseURL("https://api.test"), client.WithHTTPClient(&http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		return jsonResponse(http.StatusForbidden, `{"error":{"code":"forbidden","message":"blocked"}}`), nil
	})}))

	if _, err := sdk.GetContactDetails(context.Background(), ""); err == nil || err.Error() != "band: contact id is required" {
		t.Fatalf("GetContactDetails validation err=%v", err)
	}
	if _, err := sdk.CheckContactAvailability(context.Background(), client.CheckContactAvailabilityInput{}); err == nil || err.Error() != "band: at least one contact id is required" {
		t.Fatalf("CheckContactAvailability validation err=%v", err)
	}
	contactIDs := make([]string, 51)
	for i := range contactIDs {
		contactIDs[i] = "contact"
	}
	if _, err := sdk.CheckContactAvailability(context.Background(), client.CheckContactAvailabilityInput{ContactIDs: contactIDs}); err == nil || err.Error() != "band: at most 50 contact ids are allowed" {
		t.Fatalf("CheckContactAvailability max err=%v", err)
	}
	if _, err := sdk.SearchContacts(context.Background(), client.SearchContactsInput{}); !errors.Is(err, client.ErrForbidden) {
		t.Fatalf("SearchContacts err=%v", err)
	}
}
