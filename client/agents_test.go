package client_test

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/darvell-thenvoi/overture-thenvoi-sdk-go/client"
)

func TestGetAgent(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		apiKey        string
		responseCode  int
		responseBody  string
		assertRequest func(*testing.T, *http.Request)
		assertResult  func(*testing.T, *client.AgentIdentity, error)
		wantCalled    bool
	}{
		{
			name:         "decodes generated data envelope and sends expected request",
			apiKey:       "test-key",
			responseCode: http.StatusOK,
			responseBody: `{"data":{"id":"agent_123","name":"Band Agent","description":"handles workflows","handle":"owner/agent","inserted_at":"2026-01-02T03:04:05Z","listed_in_directory":true,"owner_uuid":"owner_1","tags":["ops"],"updated_at":"2026-01-03T03:04:05Z"}}`,
			assertRequest: func(t *testing.T, req *http.Request) {
				t.Helper()
				if req.Method != http.MethodGet {
					t.Fatalf("method = %s", req.Method)
				}
				if req.URL.String() != "https://api.test/api/v1/agent/me" {
					t.Fatalf("url = %s", req.URL.String())
				}
				if got := req.Header.Get("X-API-Key"); got != "test-key" {
					t.Fatalf("x-api-key = %s", got)
				}
			},
			assertResult: func(t *testing.T, out *client.AgentIdentity, err error) {
				t.Helper()
				if err != nil {
					t.Fatalf("GetAgent returned error: %v", err)
				}
				if out == nil || out.ID != "agent_123" || out.Name != "Band Agent" || out.Handle != "owner/agent" {
					t.Fatalf("identity = %#v", out)
				}
				if out.Description == nil || *out.Description != "handles workflows" {
					t.Fatalf("description = %#v", out.Description)
				}
			},
			wantCalled: true,
		},
		{
			name:         "maps unauthorized response",
			apiKey:       "test-key",
			responseCode: http.StatusUnauthorized,
			responseBody: `{"error":{"code":"unauthorized","message":"bad key"}}`,
			assertResult: func(t *testing.T, out *client.AgentIdentity, err error) {
				if out != nil || !errors.Is(err, client.ErrUnauthorized) {
					t.Fatalf("out=%#v err=%v", out, err)
				}
			},
			wantCalled: true,
		},
		{
			name:   "returns missing api key without request",
			apiKey: "",
			assertResult: func(t *testing.T, out *client.AgentIdentity, err error) {
				if out != nil || err == nil {
					t.Fatalf("out=%#v err=%v", out, err)
				}
			},
			wantCalled: false,
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
			sdk := client.New(client.WithApiKey(tt.apiKey), client.WithBaseURL("https://api.test"), client.WithHTTPClient(&http.Client{Transport: transport}))
			out, err := sdk.GetAgent(context.Background())
			if tt.assertResult != nil {
				tt.assertResult(t, out, err)
			}
			if called != tt.wantCalled {
				t.Fatalf("transport called = %t, want %t", called, tt.wantCalled)
			}
		})
	}
}

func TestGetAgentMe(t *testing.T) {
	t.Parallel()
	transport := roundTripFunc(func(req *http.Request) (*http.Response, error) {
		if req.URL.String() != "https://api.test/api/v1/agent/me" {
			t.Fatalf("url = %s", req.URL.String())
		}
		return jsonResponse(http.StatusOK, `{"data":{"id":"agent_789","name":"Compat","description":null,"handle":"owner/compat","inserted_at":"2026-01-02T03:04:05Z","owner_uuid":"owner_1","updated_at":"2026-01-02T03:05:05Z"}}`), nil
	})
	sdk := client.New(client.WithApiKey("test-key"), client.WithBaseURL("https://api.test"), client.WithHTTPClient(&http.Client{Transport: transport}))
	out, err := sdk.GetAgentMe(context.Background())
	if err != nil || out == nil || out.ID != "agent_789" {
		t.Fatalf("out=%#v err=%v", out, err)
	}
}
