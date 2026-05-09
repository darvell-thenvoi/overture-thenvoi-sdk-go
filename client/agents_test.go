package client_test

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/darvell-thenvoi/overture-thenvoi-sdk-go/client"
)

func TestGetAgentMe(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		apiKey        string
		responseCode  int
		responseBody  string
		assertRequest func(*testing.T, *http.Request)
		assertResult  func(*testing.T, *client.AgentIdentity, error)
	}{
		{
			name:         "decodes success and sends expected request",
			apiKey:       "test-key",
			responseCode: http.StatusOK,
			responseBody: `{"id":"agent_123","name":"Thenvoi Agent","description":"handles workflows"}`,
			assertRequest: func(t *testing.T, req *http.Request) {
				t.Helper()
				if req.Method != http.MethodGet {
					t.Fatalf("method = %s, want %s", req.Method, http.MethodGet)
				}
				if req.URL.String() != "https://api.test/v1/agents/me" {
					t.Fatalf("url = %s", req.URL.String())
				}
				if got := req.Header.Get("Authorization"); got != "Bearer test-key" {
					t.Fatalf("authorization = %s", got)
				}
			},
			assertResult: func(t *testing.T, out *client.AgentIdentity, err error) {
				t.Helper()
				if err != nil {
					t.Fatalf("GetAgentMe returned error: %v", err)
				}
				if out == nil {
					t.Fatal("GetAgentMe returned nil identity")
				}
				if out.ID != "agent_123" {
					t.Fatalf("id = %s", out.ID)
				}
				if out.Name != "Thenvoi Agent" {
					t.Fatalf("name = %s", out.Name)
				}
				if out.Description == nil || *out.Description != "handles workflows" {
					t.Fatalf("description = %#v", out.Description)
				}
			},
		},
		{
			name:         "decodes null description as nil",
			apiKey:       "test-key",
			responseCode: http.StatusOK,
			responseBody: `{"id":"agent_456","name":"No Description","description":null}`,
			assertRequest: func(t *testing.T, req *http.Request) {
				t.Helper()
				if req.URL.String() != "https://api.test/v1/agents/me" {
					t.Fatalf("url = %s", req.URL.String())
				}
			},
			assertResult: func(t *testing.T, out *client.AgentIdentity, err error) {
				t.Helper()
				if err != nil {
					t.Fatalf("GetAgentMe returned error: %v", err)
				}
				if out == nil {
					t.Fatal("GetAgentMe returned nil identity")
				}
				if out.Description != nil {
					t.Fatalf("description = %#v, want nil", out.Description)
				}
			},
		},
		{
			name:         "maps unauthorized response",
			apiKey:       "test-key",
			responseCode: http.StatusUnauthorized,
			responseBody: `{"error":{"code":"unauthorized","message":"bad key"}}`,
			assertResult: func(t *testing.T, out *client.AgentIdentity, err error) {
				t.Helper()
				if out != nil {
					t.Fatalf("identity = %#v, want nil", out)
				}
				if !errors.Is(err, client.ErrUnauthorized) {
					t.Fatalf("err = %v, want ErrUnauthorized", err)
				}
				var apiErr *client.ApiError
				if !errors.As(err, &apiErr) {
					t.Fatalf("err = %T, want *client.ApiError", err)
				}
			},
		},
		{
			name:   "returns missing api key without request",
			apiKey: "",
			assertResult: func(t *testing.T, out *client.AgentIdentity, err error) {
				t.Helper()
				if out != nil {
					t.Fatalf("identity = %#v, want nil", out)
				}
				if err == nil {
					t.Fatal("expected error, got nil")
				}
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			called := false
			transport := roundTripFunc(func(req *http.Request) (*http.Response, error) {
				called = true
				if tt.assertRequest != nil {
					tt.assertRequest(t, req)
				}
				return jsonResponse(tt.responseCode, tt.responseBody), nil
			})

			sdk := client.New(
				client.WithApiKey(tt.apiKey),
				client.WithBaseURL("https://api.test"),
				client.WithHTTPClient(&http.Client{Transport: transport}),
			)

			out, err := sdk.GetAgentMe(context.Background())
			tt.assertResult(t, out, err)

			if tt.apiKey == "" && called {
				t.Fatal("transport was called for missing api key")
			}
		})
	}
}
