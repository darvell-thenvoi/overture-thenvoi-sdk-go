package client_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/darvell-thenvoi/overture-thenvoi-sdk-go/client"
)

func TestSendChatMessage(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		chatID        string
		input         client.SendChatMessageInput
		responseCode  int
		responseBody  string
		assertRequest func(*testing.T, *http.Request)
		assertResult  func(*testing.T, *client.ChatMessage, error)
		wantCalled    bool
	}{
		{
			name:   "sends expected payload and decodes response",
			chatID: "chat_123",
			input: client.SendChatMessageInput{
				Content:     "hello room",
				MessageType: "task",
				Metadata: map[string]any{
					"priority": "high",
				},
				Mentions: []client.Mention{{
					ID:     "user_1",
					Handle: "@alice",
				}},
			},
			responseCode: http.StatusOK,
			responseBody: `{"data":{"id":"msg_1","content":"hello room","sender_id":"agent_1","sender_type":"agent","sender_name":"Thenvoi Bot","message_type":"task","metadata":{"priority":"high"},"inserted_at":"2026-01-02T03:04:05Z","updated_at":"2026-01-02T03:05:05Z"}}`,
			assertRequest: func(t *testing.T, req *http.Request) {
				t.Helper()
				if req.Method != http.MethodPost {
					t.Fatalf("method = %s, want %s", req.Method, http.MethodPost)
				}
				if req.URL.String() != "https://api.test/v1/chats/chat_123/messages" {
					t.Fatalf("url = %s", req.URL.String())
				}
				if got := req.Header.Get("Authorization"); got != "Bearer test-key" {
					t.Fatalf("authorization = %s", got)
				}
				if got := req.Header.Get("Content-Type"); got != "application/json" {
					t.Fatalf("content-type = %s", got)
				}

				var payload map[string]any
				if err := json.NewDecoder(req.Body).Decode(&payload); err != nil {
					t.Fatalf("decode request body: %v", err)
				}
				message, ok := payload["message"].(map[string]any)
				if !ok {
					t.Fatalf("message wrapper missing: %#v", payload)
				}
				if message["content"] != "hello room" {
					t.Fatalf("content = %#v", message["content"])
				}
				if message["message_type"] != "task" {
					t.Fatalf("message_type = %#v", message["message_type"])
				}
				metadata, ok := message["metadata"].(map[string]any)
				if !ok {
					t.Fatalf("metadata = %#v", message["metadata"])
				}
				if metadata["priority"] != "high" {
					t.Fatalf("priority = %#v", metadata["priority"])
				}
				mentions, ok := message["mentions"].([]any)
				if !ok || len(mentions) != 1 {
					t.Fatalf("mentions = %#v", message["mentions"])
				}
				mention, ok := mentions[0].(map[string]any)
				if !ok {
					t.Fatalf("mention = %#v", mentions[0])
				}
				if mention["id"] != "user_1" {
					t.Fatalf("mention id = %#v", mention["id"])
				}
				if mention["handle"] != "@alice" {
					t.Fatalf("mention handle = %#v", mention["handle"])
				}
				if _, exists := mention["name"]; exists {
					t.Fatalf("mention name should be omitted: %#v", mention)
				}
				if _, exists := mention["username"]; exists {
					t.Fatalf("mention username should be omitted: %#v", mention)
				}
			},
			assertResult: func(t *testing.T, out *client.ChatMessage, err error) {
				t.Helper()
				if err != nil {
					t.Fatalf("SendChatMessage returned error: %v", err)
				}
				if out == nil {
					t.Fatal("SendChatMessage returned nil message")
				}
				if out.ID != "msg_1" {
					t.Fatalf("id = %s", out.ID)
				}
				if out.SenderName == nil || *out.SenderName != "Thenvoi Bot" {
					t.Fatalf("sender_name = %#v", out.SenderName)
				}
				wantInserted, _ := time.Parse(time.RFC3339, "2026-01-02T03:04:05Z")
				if !out.InsertedAt.Equal(wantInserted) {
					t.Fatalf("inserted_at = %s", out.InsertedAt)
				}
				wantUpdated, _ := time.Parse(time.RFC3339, "2026-01-02T03:05:05Z")
				if out.UpdatedAt == nil || !out.UpdatedAt.Equal(wantUpdated) {
					t.Fatalf("updated_at = %#v", out.UpdatedAt)
				}
			},
			wantCalled: true,
		},
		{
			name:   "decodes null sender name and omitted updated at",
			chatID: "chat_123",
			input: client.SendChatMessageInput{
				Content: "hello room",
			},
			responseCode: http.StatusOK,
			responseBody: `{"data":{"id":"msg_2","content":"hello room","sender_id":"agent_1","sender_type":"agent","sender_name":null,"message_type":"task","inserted_at":"2026-01-02T03:04:05Z"}}`,
			assertResult: func(t *testing.T, out *client.ChatMessage, err error) {
				t.Helper()
				if err != nil {
					t.Fatalf("SendChatMessage returned error: %v", err)
				}
				if out.SenderName != nil {
					t.Fatalf("sender_name = %#v, want nil", out.SenderName)
				}
				if out.UpdatedAt != nil {
					t.Fatalf("updated_at = %#v, want nil", out.UpdatedAt)
				}
			},
			wantCalled: true,
		},
		{
			name:   "escapes chat id in request path",
			chatID: "a/b",
			input: client.SendChatMessageInput{
				Content: "hello",
			},
			responseCode: http.StatusOK,
			responseBody: `{"data":{"id":"msg_3","content":"hello","sender_id":"agent_1","sender_type":"agent","message_type":"task","inserted_at":"2026-01-02T03:04:05Z"}}`,
			assertRequest: func(t *testing.T, req *http.Request) {
				t.Helper()
				if req.URL.String() != "https://api.test/v1/chats/a%2Fb/messages" {
					t.Fatalf("url = %s", req.URL.String())
				}
			},
			assertResult: func(t *testing.T, out *client.ChatMessage, err error) {
				t.Helper()
				if err != nil {
					t.Fatalf("SendChatMessage returned error: %v", err)
				}
				if out == nil {
					t.Fatal("SendChatMessage returned nil message")
				}
			},
			wantCalled: true,
		},
		{
			name:   "returns validation error for empty chat id without network",
			chatID: "",
			input: client.SendChatMessageInput{
				Content: "hello",
			},
			assertResult: func(t *testing.T, out *client.ChatMessage, err error) {
				t.Helper()
				if out != nil {
					t.Fatalf("message = %#v, want nil", out)
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
			name:   "returns validation error for empty content without network",
			chatID: "chat_123",
			input:  client.SendChatMessageInput{},
			assertResult: func(t *testing.T, out *client.ChatMessage, err error) {
				t.Helper()
				if out != nil {
					t.Fatalf("message = %#v, want nil", out)
				}
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if err.Error() != "thenvoi: content is required" {
					t.Fatalf("err = %v", err)
				}
			},
			wantCalled: false,
		},
		{
			name:   "maps unauthorized response",
			chatID: "chat_123",
			input: client.SendChatMessageInput{
				Content: "hello",
			},
			responseCode: http.StatusUnauthorized,
			responseBody: `{"error":{"code":"unauthorized","message":"bad key"}}`,
			assertResult: func(t *testing.T, out *client.ChatMessage, err error) {
				t.Helper()
				if out != nil {
					t.Fatalf("message = %#v, want nil", out)
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

			out, err := sdk.SendChatMessage(context.Background(), tt.chatID, tt.input)
			if tt.assertResult != nil {
				tt.assertResult(t, out, err)
			}

			if called != tt.wantCalled {
				t.Fatalf("transport called = %t, want %t", called, tt.wantCalled)
			}
		})
	}
}
