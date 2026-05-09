package client_test

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/darvell-thenvoi/overture-thenvoi-sdk-go/client"
)

func TestGetChatRoom(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		chatID        string
		responseCode  int
		responseBody  string
		assertRequest func(*testing.T, *http.Request)
		assertResult  func(*testing.T, *client.ChatRoom, error)
		wantCalled    bool
	}{
		{
			name:         "decodes success response",
			chatID:       "chat_123",
			responseCode: http.StatusOK,
			responseBody: `{"data":{"id":"chat_123","name":"General","description":"Shared channel","type":"group","status":"active"}}`,
			assertRequest: func(t *testing.T, req *http.Request) {
				t.Helper()
				if req.Method != http.MethodGet {
					t.Fatalf("method = %s, want %s", req.Method, http.MethodGet)
				}
				if req.URL.String() != "https://api.test/api/v1/agent/chats/chat_123" {
					t.Fatalf("url = %s", req.URL.String())
				}
				if got := req.Header.Get("X-API-Key"); got != "test-key" {
					t.Fatalf("x-api-key = %s", got)
				}
			},
			assertResult: func(t *testing.T, out *client.ChatRoom, err error) {
				t.Helper()
				if err != nil {
					t.Fatalf("GetChatRoom returned error: %v", err)
				}
				if out == nil {
					t.Fatal("GetChatRoom returned nil chat room")
				}
				if out.ID != "chat_123" {
					t.Fatalf("id = %s", out.ID)
				}
				if out.Name != "General" {
					t.Fatalf("name = %s", out.Name)
				}
				if out.Description == nil || *out.Description != "Shared channel" {
					t.Fatalf("description = %#v", out.Description)
				}
				if out.Type == nil || *out.Type != "group" {
					t.Fatalf("type = %#v", out.Type)
				}
				if out.Status == nil || *out.Status != "active" {
					t.Fatalf("status = %#v", out.Status)
				}
			},
			wantCalled: true,
		},
		{
			name:         "decodes omitted optional fields",
			chatID:       "chat_456",
			responseCode: http.StatusOK,
			responseBody: `{"data":{"id":"chat_456","name":"No Extras"}}`,
			assertResult: func(t *testing.T, out *client.ChatRoom, err error) {
				t.Helper()
				if err != nil {
					t.Fatalf("GetChatRoom returned error: %v", err)
				}
				if out == nil {
					t.Fatal("GetChatRoom returned nil chat room")
				}
				if out.Description != nil {
					t.Fatalf("description = %#v, want nil", out.Description)
				}
				if out.Type != nil {
					t.Fatalf("type = %#v, want nil", out.Type)
				}
				if out.Status != nil {
					t.Fatalf("status = %#v, want nil", out.Status)
				}
			},
			wantCalled: true,
		},
		{
			name:         "escapes chat id in request path",
			chatID:       "a/b",
			responseCode: http.StatusOK,
			responseBody: `{"data":{"id":"a/b","name":"Escaped"}}`,
			assertRequest: func(t *testing.T, req *http.Request) {
				t.Helper()
				if req.URL.String() != "https://api.test/api/v1/agent/chats/a%2Fb" {
					t.Fatalf("url = %s", req.URL.String())
				}
			},
			assertResult: func(t *testing.T, out *client.ChatRoom, err error) {
				t.Helper()
				if err != nil {
					t.Fatalf("GetChatRoom returned error: %v", err)
				}
				if out == nil {
					t.Fatal("GetChatRoom returned nil chat room")
				}
			},
			wantCalled: true,
		},
		{
			name:   "returns validation error for empty chat id without network",
			chatID: "",
			assertResult: func(t *testing.T, out *client.ChatRoom, err error) {
				t.Helper()
				if out != nil {
					t.Fatalf("chat room = %#v, want nil", out)
				}
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if err.Error() != "thenvoi: chat id is required" {
					t.Fatalf("err = %v", err)
				}
			},
			wantCalled: false,
		},
		{
			name:         "maps unauthorized response",
			chatID:       "chat_123",
			responseCode: http.StatusUnauthorized,
			responseBody: `{"error":{"code":"unauthorized","message":"bad key"}}`,
			assertResult: func(t *testing.T, out *client.ChatRoom, err error) {
				t.Helper()
				if out != nil {
					t.Fatalf("chat room = %#v, want nil", out)
				}
				if !errors.Is(err, client.ErrUnauthorized) {
					t.Fatalf("err = %v, want ErrUnauthorized", err)
				}
				var apiErr *client.ApiError
				if !errors.As(err, &apiErr) {
					t.Fatalf("err = %T, want *client.ApiError", err)
				}
			},
			wantCalled: true,
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
				client.WithApiKey("test-key"),
				client.WithBaseURL("https://api.test"),
				client.WithHTTPClient(&http.Client{Transport: transport}),
			)

			out, err := sdk.GetChatRoom(context.Background(), tt.chatID)
			if tt.assertResult != nil {
				tt.assertResult(t, out, err)
			}

			if called != tt.wantCalled {
				t.Fatalf("transport called = %t, want %t", called, tt.wantCalled)
			}
		})
	}
}
