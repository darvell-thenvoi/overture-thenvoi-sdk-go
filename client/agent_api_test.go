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
			return jsonResponse(http.StatusOK, `{"data":[{"id":"msg_1","content":"hello","message_type":"text","sender_id":"user_1","sender_type":"User"}],"meta":{"page":2,"page_size":25}}`), nil
		case "/api/v1/agent/peers":
			return jsonResponse(http.StatusOK, `{"data":[{"id":"agent_2","handle":"owner/agent","is_contact":false,"name":"Agent","source":"registry","type":"Agent"}],"metadata":{"page":2,"page_size":25}}`), nil
		case "/api/v1/agent/contacts":
			return jsonResponse(http.StatusOK, `{"data":[{"id":"contact_1","handle":"owner/agent","inserted_at":"2026-01-02T03:04:05Z","type":"Agent"}],"metadata":{"page":2,"page_size":25}}`), nil
		case "/api/v1/agent/contacts/add":
			assertJSONField(t, req, "handle", "owner/agent")
			return jsonResponse(http.StatusOK, `{"data":{"id":"request_1","status":"pending"}}`), nil
		case "/api/v1/agent/contacts/remove":
			assertJSONField(t, req, "handle", "owner/agent")
			return jsonResponse(http.StatusOK, `{"data":{"status":"removed"}}`), nil
		case "/api/v1/agent/contacts/requests":
			return jsonResponse(http.StatusOK, `{"data":{"received":[{"id":"req_1","status":"pending"}],"sent":[]},"metadata":{"page":2,"page_size":25,"received":{"total":1,"total_pages":1},"sent":{"total":0,"total_pages":0}}}`), nil
		case "/api/v1/agent/contacts/requests/respond":
			assertJSONFields(t, req, map[string]any{"action": "approve", "request_id": "req_1"})
			return jsonResponse(http.StatusOK, `{"data":{"id":"req_1","status":"approved"}}`), nil
		case "/api/v1/agent/memories":
			if req.Method == http.MethodPost {
				assertNestedJSONFields(t, req, "memory", map[string]any{"content": "likes SUVs", "system": "long_term"})
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
	contactRequests, err := sdk.ListContactRequests(context.Background(), &client.ListContactRequestsInput{Page: &page, PageSize: &pageSize, SentStatus: &sentStatus})
	if err != nil {
		t.Fatal(err)
	}
	if contactRequests.Metadata == nil || contactRequests.Metadata.Received.Total != 1 || contactRequests.Metadata.Sent.Total != 0 {
		t.Fatalf("contact request metadata=%#v", contactRequests.Metadata)
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
	want := []string{
		"GET https://api.test/api/v1/agent/chats/chat_1/context?page=2&page_size=25",
		"GET https://api.test/api/v1/agent/peers?not_in_chat=chat_1&page=2&page_size=25",
		"GET https://api.test/api/v1/agent/contacts?page=2&page_size=25",
		"POST https://api.test/api/v1/agent/contacts/add",
		"POST https://api.test/api/v1/agent/contacts/remove",
		"GET https://api.test/api/v1/agent/contacts/requests?page=2&page_size=25&sent_status=pending",
		"POST https://api.test/api/v1/agent/contacts/requests/respond",
		"GET https://api.test/api/v1/agent/memories?content_query=prefers+SUV&page_size=25&scope=subject&segment=user&status=active&system=long_term&type=semantic",
		"POST https://api.test/api/v1/agent/memories",
		"GET https://api.test/api/v1/agent/memories/mem_1",
		"POST https://api.test/api/v1/agent/memories/mem_1/supersede",
		"POST https://api.test/api/v1/agent/memories/mem_1/archive",
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

func TestOptionalMutationDataResponses(t *testing.T) {
	t.Parallel()
	transport := roundTripFunc(func(req *http.Request) (*http.Response, error) {
		switch req.URL.Path {
		case "/api/v1/agent/contacts/add", "/api/v1/agent/contacts/remove", "/api/v1/agent/contacts/requests/respond", "/api/v1/agent/memories/mem_1/supersede", "/api/v1/agent/memories/mem_1/archive":
			return jsonResponse(http.StatusOK, `{}`), nil
		default:
			t.Fatalf("unexpected path %s", req.URL.String())
			return nil, nil
		}
	})
	sdk := client.New(client.WithApiKey("test-key"), client.WithBaseURL("https://api.test"), client.WithHTTPClient(&http.Client{Transport: transport}))
	if out, err := sdk.AddContact(context.Background(), client.AddContactInput{Handle: "owner/agent"}); err != nil || out != nil {
		t.Fatalf("AddContact out=%#v err=%v", out, err)
	}
	if out, err := sdk.RemoveContact(context.Background(), client.RemoveContactInput{Handle: stringPtr("owner/agent")}); err != nil || out != nil {
		t.Fatalf("RemoveContact out=%#v err=%v", out, err)
	}
	if out, err := sdk.RespondContactRequest(context.Background(), client.RespondContactRequestInput{Action: "approve", RequestID: stringPtr("req_1")}); err != nil || out != nil {
		t.Fatalf("RespondContactRequest out=%#v err=%v", out, err)
	}
	if out, err := sdk.SupersedeMemory(context.Background(), "mem_1"); err != nil || out != nil {
		t.Fatalf("SupersedeMemory out=%#v err=%v", out, err)
	}
	if out, err := sdk.ArchiveMemory(context.Background(), "mem_1"); err != nil || out != nil {
		t.Fatalf("ArchiveMemory out=%#v err=%v", out, err)
	}
}

func TestClientSideValidation(t *testing.T) {
	t.Parallel()
	sdk := client.New(client.WithApiKey("test-key"), client.WithHTTPClient(&http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		t.Fatalf("network should not be called for invalid input")
		return nil, nil
	})}))
	if _, err := sdk.SendChatMessage(context.Background(), "chat_1", client.SendChatMessageInput{Content: "hello", Mentions: []client.Mention{{}}}); err == nil || err.Error() != "band: mention id is required" {
		t.Fatalf("mention validation err=%v", err)
	}
	if _, err := sdk.RemoveContact(context.Background(), client.RemoveContactInput{Handle: stringPtr("")}); err == nil || err.Error() != "band: handle or contact id is required" {
		t.Fatalf("remove validation err=%v", err)
	}
	if _, err := sdk.RespondContactRequest(context.Background(), client.RespondContactRequestInput{Action: "approve"}); err == nil || err.Error() != "band: handle or request id is required" {
		t.Fatalf("respond validation err=%v", err)
	}
}

func assertJSONField(t *testing.T, req *http.Request, key string, want any) {
	t.Helper()
	assertJSONFields(t, req, map[string]any{key: want})
}

func assertJSONFields(t *testing.T, req *http.Request, want map[string]any) {
	t.Helper()
	var payload map[string]any
	decodeJSONBody(t, req, &payload)
	for key, value := range want {
		if payload[key] != value {
			t.Fatalf("payload[%s]=%#v want %#v", key, payload[key], value)
		}
	}
}

func decodeJSONBody(t *testing.T, req *http.Request, target any) {
	t.Helper()
	if err := json.NewDecoder(req.Body).Decode(target); err != nil {
		t.Fatalf("decode request body: %v", err)
	}
}

func assertNestedJSONFields(t *testing.T, req *http.Request, outer string, want map[string]any) {
	t.Helper()
	var payload map[string]map[string]any
	decodeJSONBody(t, req, &payload)
	for key, value := range want {
		if payload[outer][key] != value {
			t.Fatalf("payload[%s][%s]=%#v want %#v", outer, key, payload[outer][key], value)
		}
	}
}

func stringPtr(value string) *string {
	return &value
}
