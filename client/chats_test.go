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
		apiKey        string
		responseCode  int
		responseBody  string
		assertRequest func(*testing.T, *http.Request)
		assertResult  func(*testing.T, *client.ChatRoom, error)
		wantCalled    bool
	}{
		{
			name:         "decodes success and sends expected request",
			chatID:       "chat_123",
			apiKey:       "test-key",
			responseCode: http.StatusOK,
			responseBody: `{"data":{"id":"chat_123","inserted_at":"2026-01-02T03:04:05Z","task_id":"task_123","title":"Daily Standup","updated_at":"2026-01-02T03:06:05Z"}}`,
			assertRequest: func(t *testing.T, req *http.Request) {
				t.Helper()
				if req.Method != http.MethodGet {
					t.Fatalf("method = %s, want %s", req.Method, http.MethodGet)
				}
				if req.URL.String() != "https://api.test/v1/chats/chat_123" {
					t.Fatalf("url = %s", req.URL.String())
				}
				if got := req.Header.Get("Authorization"); got != "Bearer test-key" {
					t.Fatalf("authorization = %s", got)
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
				if out.InsertedAt != "2026-01-02T03:04:05Z" {
					t.Fatalf("inserted_at = %s", out.InsertedAt)
				}
				if out.TaskID == nil || *out.TaskID != "task_123" {
					t.Fatalf("task_id = %#v", out.TaskID)
				}
				if out.Title == nil || *out.Title != "Daily Standup" {
					t.Fatalf("title = %#v", out.Title)
				}
				if out.UpdatedAt != "2026-01-02T03:06:05Z" {
					t.Fatalf("updated_at = %s", out.UpdatedAt)
				}
			},
			wantCalled: true,
		},
		{
			name:         "decodes optional fields when omitted",
			chatID:       "chat_456",
			apiKey:       "test-key",
			responseCode: http.StatusOK,
			responseBody: `{"data":{"id":"chat_456","inserted_at":"2026-02-02T03:04:05Z","updated_at":"2026-02-02T03:06:05Z"}}`,
			assertResult: func(t *testing.T, out *client.ChatRoom, err error) {
				t.Helper()
				if err != nil {
					t.Fatalf("GetChatRoom returned error: %v", err)
				}
				if out == nil {
					t.Fatal("GetChatRoom returned nil chat room")
				}
				if out.TaskID != nil {
					t.Fatalf("task_id = %#v, want nil", out.TaskID)
				}
				if out.Title != nil {
					t.Fatalf("title = %#v, want nil", out.Title)
				}
			},
			wantCalled: true,
		},
		{
			name:         "escapes chat id in request path",
			chatID:       "a/b",
			apiKey:       "test-key",
			responseCode: http.StatusOK,
			responseBody: `{"data":{"id":"a/b","inserted_at":"2026-01-02T03:04:05Z","updated_at":"2026-01-02T03:06:05Z"}}`,
			assertRequest: func(t *testing.T, req *http.Request) {
				t.Helper()
				if req.URL.String() != "https://api.test/v1/chats/a%2Fb" {
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
			apiKey: "test-key",
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
			apiKey:       "test-key",
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
				client.WithApiKey(tt.apiKey),
				client.WithBaseURL("https://api.test"),
				client.WithHTTPClient(&http.Client{Transport: transport}),
			)

			out, err := sdk.GetChatRoom(context.Background(), tt.chatID)
			tt.assertResult(t, out, err)

			if called != tt.wantCalled {
				t.Fatalf("transport called = %t, want %t", called, tt.wantCalled)
			}
		})
	}
}
