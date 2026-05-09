package client_test

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/darvell-thenvoi/overture-thenvoi-sdk-go/client"
)

func TestParticipantEndpoints(t *testing.T) {
	t.Parallel()
	role := "admin"
	seen := []string{}
	transport := roundTripFunc(func(req *http.Request) (*http.Response, error) {
		seen = append(seen, req.Method+" "+req.URL.String())
		if req.Method == http.MethodPost {
			var payload map[string]map[string]any
			if err := json.NewDecoder(req.Body).Decode(&payload); err != nil {
				t.Fatalf("decode request body: %v", err)
			}
			participant := payload["participant"]
			if participant["participant_id"] != "agent_2" || participant["role"] != "admin" {
				t.Fatalf("participant payload = %#v", participant)
			}
		}
		if req.Method == http.MethodGet {
			return jsonResponse(http.StatusOK, `{"data":[{"id":"agent_2","handle":"owner/agent","name":"Agent","role":"admin","status":"active","type":"Agent"}]}`), nil
		}
		return jsonResponse(http.StatusOK, `{"data":{"id":"agent_2","handle":"owner/agent","name":"Agent","role":"admin","status":"active","type":"Agent"}}`), nil
	})
	sdk := client.New(client.WithApiKey("test-key"), client.WithBaseURL("https://api.test"), client.WithHTTPClient(&http.Client{Transport: transport}))
	if _, err := sdk.ListChatParticipants(context.Background(), "chat/1"); err != nil {
		t.Fatal(err)
	}
	if _, err := sdk.AddChatParticipant(context.Background(), "chat/1", client.AddChatParticipantInput{ParticipantID: "agent_2", Role: &role}); err != nil {
		t.Fatal(err)
	}
	if _, err := sdk.RemoveChatParticipant(context.Background(), "chat/1", "agent/2"); err != nil {
		t.Fatal(err)
	}
	want := []string{
		"GET https://api.test/api/v1/agent/chats/chat%2F1/participants",
		"POST https://api.test/api/v1/agent/chats/chat%2F1/participants",
		"DELETE https://api.test/api/v1/agent/chats/chat%2F1/participants/agent%2F2",
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

func TestContextPeersContactsAndMemoriesEndpoints(t *testing.T) {
	t.Parallel()
	page := 2
	pageSize := 25
	notInChat := "chat_1"
	sentStatus := "pending"
	scope := "subject"
	system := "long_term"
	memoryType := "semantic"
	segment := "user"
	status := "active"
	query := "prefers SUV"
	seen := []string{}
	transport := roundTripFunc(func(req *http.Request) (*http.Response, error) {
		seen = append(seen, req.Method+" "+req.URL.String())
		switch req.URL.Path {
		case "/api/v1/agent/chats/chat_1/context":
			return jsonResponse(http.StatusOK, `{"data":[{"id":"msg_1","content":"hello","message_type":"text","sender_id":"user_1","sender_type":"User"}],"metadata":{"page":2,"page_size":25}}`), nil
		case "/api/v1/agent/peers":
			return jsonResponse(http.StatusOK, `{"data":[{"id":"agent_2","handle":"owner/agent","is_contact":false,"name":"Agent","source":"registry","type":"Agent"}],"metadata":{"page":2,"page_size":25}}`), nil
		case "/api/v1/agent/contacts":
			return jsonResponse(http.StatusOK, `{"data":[{"id":"contact_1","handle":"owner/agent","inserted_at":"2026-01-02T03:04:05Z","type":"Agent"}],"metadata":{"page":2,"page_size":25}}`), nil
		case "/api/v1/agent/contacts/add":
			return jsonResponse(http.StatusOK, `{"data":{"id":"request_1","status":"pending"}}`), nil
		case "/api/v1/agent/contacts/remove":
			return jsonResponse(http.StatusOK, `{"data":{"status":"removed"}}`), nil
		case "/api/v1/agent/contacts/requests":
			return jsonResponse(http.StatusOK, `{"data":{"received":[{"id":"req_1","status":"pending"}],"sent":[]},"metadata":{"page":2,"page_size":25}}`), nil
		case "/api/v1/agent/contacts/requests/respond":
			return jsonResponse(http.StatusOK, `{"data":{"id":"req_1","status":"approved"}}`), nil
		case "/api/v1/agent/memories":
			if req.Method == http.MethodPost {
				return jsonResponse(http.StatusOK, `{"data":{"id":"mem_1","content":"likes SUVs","inserted_at":"2026-01-02T03:04:05Z","scope":"subject","segment":"user","system":"long_term","type":"semantic"}}`), nil
			}
			return jsonResponse(http.StatusOK, `{"data":[{"id":"mem_1","content":"likes SUVs","inserted_at":"2026-01-02T03:04:05Z","scope":"subject","segment":"user","system":"long_term","type":"semantic"}],"meta":{"page_size":25}}`), nil
		case "/api/v1/agent/memories/mem_1", "/api/v1/agent/memories/mem_1/supersede", "/api/v1/agent/memories/mem_1/archive":
			return jsonResponse(http.StatusOK, `{"data":{"id":"mem_1","content":"likes SUVs","inserted_at":"2026-01-02T03:04:05Z","scope":"subject","segment":"user","system":"long_term","type":"semantic"}}`), nil
		default:
			t.Fatalf("unexpected path %s", req.URL.String())
			return nil, nil
		}
	})
	sdk := client.New(client.WithApiKey("test-key"), client.WithBaseURL("https://api.test"), client.WithHTTPClient(&http.Client{Transport: transport}))
	if _, err := sdk.GetChatContext(context.Background(), "chat_1", &client.GetChatContextInput{Page: &page, PageSize: &pageSize}); err != nil {
		t.Fatal(err)
	}
	if _, err := sdk.ListPeers(context.Background(), &client.ListPeersInput{Page: &page, PageSize: &pageSize, NotInChat: &notInChat}); err != nil {
		t.Fatal(err)
	}
	if _, err := sdk.ListContacts(context.Background(), &client.ListContactsInput{Page: &page, PageSize: &pageSize}); err != nil {
		t.Fatal(err)
	}
	if _, err := sdk.AddContact(context.Background(), client.AddContactInput{Handle: "owner/agent"}); err != nil {
		t.Fatal(err)
	}
	if _, err := sdk.RemoveContact(context.Background(), client.RemoveContactInput{Handle: stringPtr("owner/agent")}); err != nil {
		t.Fatal(err)
	}
	if _, err := sdk.ListContactRequests(context.Background(), &client.ListContactRequestsInput{Page: &page, PageSize: &pageSize, SentStatus: &sentStatus}); err != nil {
		t.Fatal(err)
	}
	if _, err := sdk.RespondContactRequest(context.Background(), client.RespondContactRequestInput{Action: "approve", RequestID: stringPtr("req_1")}); err != nil {
		t.Fatal(err)
	}
	if _, err := sdk.ListMemories(context.Background(), &client.ListMemoriesInput{Scope: &scope, System: &system, Type: &memoryType, Segment: &segment, Status: &status, ContentQuery: &query, PageSize: &pageSize}); err != nil {
		t.Fatal(err)
	}
	if _, err := sdk.CreateMemory(context.Background(), client.CreateMemoryInput{Content: "likes SUVs", Segment: "user", System: "long_term", Thought: "user preference", Type: "semantic"}); err != nil {
		t.Fatal(err)
	}
	if _, err := sdk.GetMemory(context.Background(), "mem_1"); err != nil {
		t.Fatal(err)
	}
	if _, err := sdk.SupersedeMemory(context.Background(), "mem_1"); err != nil {
		t.Fatal(err)
	}
	if _, err := sdk.ArchiveMemory(context.Background(), "mem_1"); err != nil {
		t.Fatal(err)
	}
	wantFirst := "GET https://api.test/api/v1/agent/chats/chat_1/context?page=2&page_size=25"
	if seen[0] != wantFirst {
		t.Fatalf("seen[0]=%s want %s", seen[0], wantFirst)
	}
}

func stringPtr(value string) *string {
	return &value
}
