package client_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"testing"

	"github.com/darvell-thenvoi/overture-thenvoi-sdk-go/client"
)

func TestListTools(t *testing.T) {
	t.Parallel()

	sdk := newToolsTestClient(t, func(r *http.Request) (*http.Response, error) {
		assertRequest(t, r, http.MethodGet, "/api/v2/tools?page=2&per_page=50")
		return jsonResponse(http.StatusOK, `{"data":[{"id":"tool-1","name":"lookup","description":"Lookup","owner_uuid":"owner-1","organization_id":null,"inserted_at":"2026-01-02T03:04:05Z","updated_at":"2026-01-02T03:04:06Z"}],"pagination":{"page":2,"per_page":50,"total_pages":3,"total_items":101}}`), nil
	})

	page := 2
	perPage := 50
	got, err := sdk.ListTools(context.Background(), &client.ListToolsInput{Page: &page, PerPage: &perPage})
	if err != nil {
		t.Fatalf("ListTools returned error: %v", err)
	}
	if len(got.Data) != 1 || got.Data[0].ID != "tool-1" || got.Pagination.PerPage != 50 || got.Pagination.TotalItems != 101 {
		t.Fatalf("response = %+v", got)
	}
}

func TestCreateTool(t *testing.T) {
	t.Parallel()

	sdk := newToolsTestClient(t, func(r *http.Request) (*http.Response, error) {
		assertRequest(t, r, http.MethodPost, "/api/v2/tools")
		var body client.CreateToolInput
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		if body.Name != "lookup_user" || body.JSONSchema["type"] != "object" {
			t.Fatalf("body = %+v", body)
		}
		if body.ConnectionConfig == nil || body.ConnectionConfig.Auth.KeyName == nil || *body.ConnectionConfig.Auth.KeyName != "TOOL_TOKEN" {
			t.Fatalf("connection_config = %+v", body.ConnectionConfig)
		}
		return jsonResponse(http.StatusCreated, toolResponseJSON("tool-1")), nil
	})

	got, err := sdk.CreateTool(context.Background(), client.CreateToolInput{
		Name:             "lookup_user",
		Description:      "Lookup a user",
		JSONSchema:       map[string]any{"type": "object"},
		ConnectionConfig: bearerConnectionConfig("TOOL_TOKEN"),
	})
	if err != nil {
		t.Fatalf("CreateTool returned error: %v", err)
	}
	if got.ID != "tool-1" || got.ConnectionConfig.Auth.Type != "bearer" {
		t.Fatalf("tool = %+v", got)
	}
}

func TestGetTool(t *testing.T) {
	t.Parallel()

	sdk := newToolsTestClient(t, func(r *http.Request) (*http.Response, error) {
		assertRequest(t, r, http.MethodGet, "/api/v2/tools/tool-1")
		return jsonResponse(http.StatusOK, toolResponseJSON("tool-1")), nil
	})

	got, err := sdk.GetTool(context.Background(), "tool-1")
	if err != nil {
		t.Fatalf("GetTool returned error: %v", err)
	}
	if got.ID != "tool-1" || got.JSONSchema["type"] != "object" {
		t.Fatalf("tool = %+v", got)
	}
}

func TestUpdateTool(t *testing.T) {
	t.Parallel()

	sdk := newToolsTestClient(t, func(r *http.Request) (*http.Response, error) {
		assertRequest(t, r, http.MethodPut, "/api/v2/tools/tool-1")
		var body client.UpdateToolInput
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		if body.Name == nil || *body.Name != "lookup_account" {
			t.Fatalf("name = %v", body.Name)
		}
		if body.ConnectionConfig == nil || body.ConnectionConfig.Auth.Type != "none" {
			t.Fatalf("connection_config = %+v", body.ConnectionConfig)
		}
		return jsonResponse(http.StatusOK, toolResponseJSON("tool-1")), nil
	})

	name := "lookup_account"
	got, err := sdk.UpdateTool(context.Background(), "tool-1", client.UpdateToolInput{
		Name:             &name,
		ConnectionConfig: noAuthConnectionConfig(),
	})
	if err != nil {
		t.Fatalf("UpdateTool returned error: %v", err)
	}
	if got.ID != "tool-1" {
		t.Fatalf("tool = %+v", got)
	}
}

func TestDeleteTool(t *testing.T) {
	t.Parallel()

	sdk := newToolsTestClient(t, func(r *http.Request) (*http.Response, error) {
		assertRequest(t, r, http.MethodDelete, "/api/v2/tools/tool-1")
		return jsonResponse(http.StatusNoContent, ""), nil
	})

	if err := sdk.DeleteTool(context.Background(), "tool-1"); err != nil {
		t.Fatalf("DeleteTool returned error: %v", err)
	}
}

func TestAssignToolsToAgent(t *testing.T) {
	t.Parallel()

	sdk := newToolsTestClient(t, func(r *http.Request) (*http.Response, error) {
		assertRequest(t, r, http.MethodPost, "/api/v2/agents/agent-1/tools")
		assertToolIDsBody(t, r, []string{"tool-1", "tool-2"})
		return jsonResponse(http.StatusOK, `{"data":{"agent_id":"agent-1","assigned_tools":[{"id":"tool-1","name":"lookup"},{"id":"tool-2","name":"search"}]}}`), nil
	})

	got, err := sdk.AssignToolsToAgent(context.Background(), "agent-1", client.AssignToolsInput{ToolIDs: []string{"tool-1", "tool-2"}})
	if err != nil {
		t.Fatalf("AssignToolsToAgent returned error: %v", err)
	}
	if got.AgentID != "agent-1" || len(got.AssignedTools) != 2 || got.AssignedTools[1].Name != "search" {
		t.Fatalf("response = %+v", got)
	}
}

func TestListAgentTools(t *testing.T) {
	t.Parallel()

	sdk := newToolsTestClient(t, func(r *http.Request) (*http.Response, error) {
		assertRequest(t, r, http.MethodGet, "/api/v2/agents/agent-1/tools?page=1&per_page=25")
		return jsonResponse(http.StatusOK, `{"data":[{"id":"tool-1","name":"lookup","description":"Lookup","assigned_at":"2026-01-02T03:04:05Z"}],"pagination":{"page":1,"per_page":25,"total_pages":1,"total_items":1}}`), nil
	})

	page := 1
	perPage := 25
	got, err := sdk.ListAgentTools(context.Background(), "agent-1", &client.ListAgentToolsInput{Page: &page, PerPage: &perPage})
	if err != nil {
		t.Fatalf("ListAgentTools returned error: %v", err)
	}
	if len(got.Data) != 1 || got.Data[0].AssignedAt.IsZero() || got.Pagination.TotalItems != 1 {
		t.Fatalf("response = %+v", got)
	}
}

func TestRemoveToolFromAgent(t *testing.T) {
	t.Parallel()

	sdk := newToolsTestClient(t, func(r *http.Request) (*http.Response, error) {
		assertRequest(t, r, http.MethodDelete, "/api/v2/agents/agent-1/tools/tool-1")
		return jsonResponse(http.StatusNoContent, ""), nil
	})

	if err := sdk.RemoveToolFromAgent(context.Background(), "agent-1", "tool-1"); err != nil {
		t.Fatalf("RemoveToolFromAgent returned error: %v", err)
	}
}

func TestListMyTools(t *testing.T) {
	t.Parallel()

	sdk := newToolsTestClient(t, func(r *http.Request) (*http.Response, error) {
		assertRequest(t, r, http.MethodGet, "/api/v2/agents/me/tools?page=3&per_page=10")
		return jsonResponse(http.StatusOK, `{"data":[{"id":"tool-1","name":"lookup","description":"Lookup","json_schema":{"type":"object"},"assigned_at":"2026-01-02T03:04:05Z"}],"pagination":{"page":3,"per_page":10,"total_pages":4,"total_items":31}}`), nil
	})

	page := 3
	perPage := 10
	got, err := sdk.ListMyTools(context.Background(), &client.ListMyToolsInput{Page: &page, PerPage: &perPage})
	if err != nil {
		t.Fatalf("ListMyTools returned error: %v", err)
	}
	if len(got.Data) != 1 || got.Data[0].JSONSchema["type"] != "object" || got.Pagination.Page != 3 {
		t.Fatalf("response = %+v", got)
	}
}

func TestAssignToolsToMyself(t *testing.T) {
	t.Parallel()

	sdk := newToolsTestClient(t, func(r *http.Request) (*http.Response, error) {
		assertRequest(t, r, http.MethodPost, "/api/v2/agents/me/tools")
		assertToolIDsBody(t, r, []string{"tool-1"})
		return jsonResponse(http.StatusOK, `{"data":{"assigned_tools":[{"id":"tool-1","name":"lookup"}]}}`), nil
	})

	got, err := sdk.AssignToolsToMyself(context.Background(), client.AssignToolsInput{ToolIDs: []string{"tool-1"}})
	if err != nil {
		t.Fatalf("AssignToolsToMyself returned error: %v", err)
	}
	if len(got.AssignedTools) != 1 || got.AssignedTools[0].ID != "tool-1" {
		t.Fatalf("response = %+v", got)
	}
}

func TestRemoveToolFromMyself(t *testing.T) {
	t.Parallel()

	sdk := newToolsTestClient(t, func(r *http.Request) (*http.Response, error) {
		assertRequest(t, r, http.MethodDelete, "/api/v2/agents/me/tools/tool-1")
		return jsonResponse(http.StatusNoContent, ""), nil
	})

	if err := sdk.RemoveToolFromMyself(context.Background(), "tool-1"); err != nil {
		t.Fatalf("RemoveToolFromMyself returned error: %v", err)
	}
}

func TestToolMethodsPropagateAPIError(t *testing.T) {
	t.Parallel()

	sdk := newToolsTestClient(t, func(r *http.Request) (*http.Response, error) {
		assertRequest(t, r, http.MethodGet, "/api/v2/tools/tool-1")
		return jsonResponse(http.StatusUnauthorized, `{"error":{"code":"unauthorized","message":"bad key"}}`), nil
	})

	_, err := sdk.GetTool(context.Background(), "tool-1")
	if !errors.Is(err, client.ErrUnauthorized) {
		t.Fatalf("err = %v, want ErrUnauthorized", err)
	}
}

func TestToolMethodsValidateRequiredInput(t *testing.T) {
	t.Parallel()

	sdk := newToolsTestClient(t, func(r *http.Request) (*http.Response, error) {
		t.Fatal("transport should not be called")
		return nil, nil
	})

	if _, err := sdk.GetTool(context.Background(), ""); err == nil {
		t.Fatal("GetTool returned nil error")
	}
	if _, err := sdk.AssignToolsToAgent(context.Background(), "", client.AssignToolsInput{ToolIDs: []string{"tool-1"}}); err == nil {
		t.Fatal("AssignToolsToAgent returned nil error")
	}
	if _, err := sdk.AssignToolsToMyself(context.Background(), client.AssignToolsInput{}); err == nil {
		t.Fatal("AssignToolsToMyself returned nil error")
	}
	if _, err := sdk.CreateTool(context.Background(), client.CreateToolInput{}); err == nil {
		t.Fatal("CreateTool returned nil error")
	}
}

func newToolsTestClient(t *testing.T, fn roundTripFunc) *client.Client {
	t.Helper()
	return client.New(
		client.WithApiKey("test-key"),
		client.WithBaseURL("https://api.test"),
		client.WithHTTPClient(&http.Client{Transport: fn}),
	)
}

func assertRequest(t *testing.T, r *http.Request, method string, pathWithQuery string) {
	t.Helper()
	if r.Method != method {
		t.Fatalf("method = %s, want %s", r.Method, method)
	}
	if got := r.URL.RequestURI(); got != pathWithQuery {
		t.Fatalf("request uri = %s, want %s", got, pathWithQuery)
	}
}

func assertToolIDsBody(t *testing.T, r *http.Request, want []string) {
	t.Helper()
	var body client.AssignToolsInput
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if len(body.ToolIDs) != len(want) {
		t.Fatalf("tool_ids = %v, want %v", body.ToolIDs, want)
	}
	for i := range want {
		if body.ToolIDs[i] != want[i] {
			t.Fatalf("tool_ids = %v, want %v", body.ToolIDs, want)
		}
	}
}

func bearerConnectionConfig(keyName string) *client.ConnectionConfig {
	return &client.ConnectionConfig{
		BaseURL:   "https://tools.test",
		Method:    "POST",
		Path:      "/lookup",
		ParamType: "body",
		Auth: client.AuthConfig{
			Type:    "bearer",
			KeyName: &keyName,
		},
	}
}

func noAuthConnectionConfig() *client.ConnectionConfig {
	return &client.ConnectionConfig{
		BaseURL:   "https://tools.test",
		Method:    "GET",
		Path:      "/lookup",
		ParamType: "query",
		Auth:      client.AuthConfig{Type: "none"},
	}
}

func toolResponseJSON(id string) string {
	return `{"data":{"id":"` + id + `","name":"lookup_user","description":"Lookup a user","json_schema":{"type":"object"},"connection_config":{"base_url":"https://tools.test","method":"POST","path":"/lookup","param_type":"body","auth":{"type":"bearer","key_name":"TOOL_TOKEN"}},"owner_uuid":"owner-1","organization_id":null,"inserted_at":"2026-01-02T03:04:05Z","updated_at":"2026-01-02T03:04:06Z"}}`
}
