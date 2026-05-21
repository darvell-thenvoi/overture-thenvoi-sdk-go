package client_test

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/darvell-thenvoi/overture-thenvoi-sdk-go/client"
)

func TestAgentAdminEndpoints(t *testing.T) {
	t.Parallel()
	page := 2
	perPage := 25
	seen := []string{}
	transport := roundTripFunc(func(req *http.Request) (*http.Response, error) {
		seen = append(seen, req.Method+" "+req.URL.String())
		switch req.URL.EscapedPath() {
		case "/api/v2/agents":
			if req.Method == http.MethodPost {
				assertJSONFields(t, req, map[string]any{"name": "Helper", "description": "Builds useful things", "model_type": "gpt-4"})
				return jsonResponse(http.StatusCreated, `{"data":{"id":"agent_1","name":"Helper","description":"Builds useful things","model_type":"gpt-4","is_external":false,"owner_uuid":"owner_1","organization_id":null,"system_prompt_id":null,"structured_output_schema":{"type":"object"},"inserted_at":"2026-01-02T03:04:05Z","updated_at":"2026-01-03T03:04:05Z"}}`), nil
			}
			return jsonResponse(http.StatusOK, `{"data":[{"id":"agent_1","name":"Helper","description":"Builds useful things","model_type":"gpt-4","is_external":false,"owner_uuid":"owner_1","organization_id":null,"system_prompt_id":null,"inserted_at":"2026-01-02T03:04:05Z","updated_at":"2026-01-03T03:04:05Z"}],"pagination":{"page":2,"per_page":25,"total_pages":1,"total_items":1}}`), nil
		case "/api/v2/agents/register":
			assertJSONFields(t, req, map[string]any{"name": "External", "description": "Runs outside Band", "model_type": "custom"})
			return jsonResponse(http.StatusCreated, `{"data":{"agent":{"id":"agent_2","name":"External","description":"Runs outside Band","model_type":"custom","is_external":true,"owner_uuid":"owner_1","organization_id":null,"system_prompt_id":null,"structured_output_schema":null,"inserted_at":"2026-01-02T03:04:05Z","updated_at":"2026-01-03T03:04:05Z"},"credentials":{"api_key":"key_1","message":"store this key"}}}`), nil
		case "/api/v2/agents/agent%2F1":
			if req.Method == http.MethodDelete {
				return jsonResponse(http.StatusNoContent, ``), nil
			}
			assertJSONField(t, req, "description", "Updated description")
			return jsonResponse(http.StatusOK, `{"data":{"id":"agent/1","name":"Helper","description":"Updated description","model_type":"gpt-4","is_external":false,"owner_uuid":"owner_1","organization_id":null,"system_prompt_id":null,"structured_output_schema":null,"inserted_at":"2026-01-02T03:04:05Z","updated_at":"2026-01-03T03:04:05Z"}}`), nil
		case "/api/v2/agents/agent%2F1/execute":
			assertJSONField(t, req, "request", "Summarize this")
			return jsonResponse(http.StatusCreated, `{"data":{"execution_id":"exec_1","task_id":"task_1","chat_room_id":"chat_1","status":"queued","agent":{"id":"agent_1","name":"Helper"},"request":"Summarize this","created_at":"2026-01-02T03:04:05Z","links":{"chat_room":"/chats/chat_1","messages":"/chats/chat_1/messages","task":"/tasks/task_1"}}}`), nil
		case "/api/v2/agents/me":
			assertJSONField(t, req, "name", "Renamed")
			return jsonResponse(http.StatusOK, `{"data":{"id":"agent_1","name":"Renamed","description":"Builds useful things","model_type":"gpt-4","is_external":false,"owner_uuid":"owner_1","organization_id":null,"system_prompt_id":null,"structured_output_schema":null,"inserted_at":"2026-01-02T03:04:05Z","updated_at":"2026-01-03T03:04:05Z"}}`), nil
		case "/api/v2/agents/me/chats":
			return jsonResponse(http.StatusOK, `{"data":[{"id":"chat_1","title":"Ops","status":"active","inserted_at":"2026-01-02T03:04:05Z","updated_at":"2026-01-03T03:04:05Z"}],"pagination":{"page":2,"per_page":25,"total_pages":1,"total_items":1}}`), nil
		default:
			t.Fatalf("unexpected path %s", req.URL.String())
			return nil, nil
		}
	})
	sdk := client.New(client.WithApiKey("test-key"), client.WithBaseURL("https://api.test"), client.WithHTTPClient(&http.Client{Transport: transport}))

	agents, err := sdk.ListAgents(context.Background(), &client.ListAgentsInput{Page: &page, PerPage: &perPage})
	if err != nil || len(agents.Data) != 1 || agents.Pagination.TotalItems != 1 {
		t.Fatalf("ListAgents out=%#v err=%v", agents, err)
	}
	registered, err := sdk.RegisterExternalAgent(context.Background(), client.RegisterExternalAgentInput{Name: "External", Description: "Runs outside Band", ModelType: "custom"})
	if err != nil || registered.Credentials.APIKey != "key_1" {
		t.Fatalf("RegisterExternalAgent out=%#v err=%v", registered, err)
	}
	created, err := sdk.CreatePlatformAgent(context.Background(), client.CreatePlatformAgentInput{Name: "Helper", Description: "Builds useful things", ModelType: "gpt-4", StructuredOutputSchema: map[string]any{"type": "object"}})
	if err != nil || created.Name != "Helper" {
		t.Fatalf("CreatePlatformAgent out=%#v err=%v", created, err)
	}
	if _, err := sdk.UpdateAgent(context.Background(), "agent/1", client.UpdateAgentInput{Description: "Updated description"}); err != nil {
		t.Fatalf("UpdateAgent returned error: %v", err)
	}
	if err := sdk.DeleteAgent(context.Background(), "agent/1"); err != nil {
		t.Fatalf("DeleteAgent returned error: %v", err)
	}
	executed, err := sdk.ExecuteAgent(context.Background(), "agent/1", client.ExecuteAgentInput{Request: "Summarize this"})
	if err != nil || executed.Links.Task != "/tasks/task_1" {
		t.Fatalf("ExecuteAgent out=%#v err=%v", executed, err)
	}
	if _, err := sdk.UpdateCurrentAgent(context.Background(), client.UpdateCurrentAgentInput{Name: "Renamed"}); err != nil {
		t.Fatalf("UpdateCurrentAgent returned error: %v", err)
	}
	chats, err := sdk.ListAgentChatRooms(context.Background(), &client.ListAgentChatRoomsInput{Page: &page, PerPage: &perPage})
	if err != nil || len(chats.Data) != 1 {
		t.Fatalf("ListAgentChatRooms out=%#v err=%v", chats, err)
	}

	want := []string{
		"GET https://api.test/api/v2/agents?page=2&per_page=25",
		"POST https://api.test/api/v2/agents/register",
		"POST https://api.test/api/v2/agents",
		"PUT https://api.test/api/v2/agents/agent%2F1",
		"DELETE https://api.test/api/v2/agents/agent%2F1",
		"POST https://api.test/api/v2/agents/agent%2F1/execute",
		"PUT https://api.test/api/v2/agents/me",
		"GET https://api.test/api/v2/agents/me/chats?page=2&per_page=25",
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

func TestAgentAdminValidationAndAPIErrors(t *testing.T) {
	t.Parallel()
	called := false
	sdk := client.New(client.WithApiKey("test-key"), client.WithBaseURL("https://api.test"), client.WithHTTPClient(&http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		called = true
		return jsonResponse(http.StatusUnauthorized, `{"error":{"code":"unauthorized","message":"bad key"}}`), nil
	})}))

	if _, err := sdk.CreatePlatformAgent(context.Background(), client.CreatePlatformAgentInput{}); err == nil || err.Error() != "band: name is required" {
		t.Fatalf("CreatePlatformAgent validation err=%v", err)
	}
	if _, err := sdk.UpdateAgent(context.Background(), "", client.UpdateAgentInput{}); err == nil || err.Error() != "band: agent id is required" {
		t.Fatalf("UpdateAgent validation err=%v", err)
	}
	if err := sdk.DeleteAgent(context.Background(), "agent_1"); !errors.Is(err, client.ErrUnauthorized) {
		t.Fatalf("DeleteAgent err=%v", err)
	}
	if !called {
		t.Fatal("transport was not called for API error check")
	}
}
