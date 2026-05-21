package client_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"testing"

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
		assertResult  func(*testing.T, *client.MessageSentResponse, error)
		wantCalled    bool
	}{
		{
			name:   "sends generated text payload and decodes response",
			chatID: "chat_123",
			input: client.SendChatMessageInput{
				Content:  "@alice hello room",
				Mentions: []client.Mention{{ID: "user_1", Handle: "alice", Name: "Alice"}},
			},
			responseCode: http.StatusOK,
			responseBody: `{"data":{"id":"msg_1","success":true,"recipients":[{"id":"user_1","handle":"alice","name":"Alice"}]}}`,
			assertRequest: func(t *testing.T, req *http.Request) {
				t.Helper()
				if req.Method != http.MethodPost {
					t.Fatalf("method = %s", req.Method)
				}
				if req.URL.String() != "https://api.test/api/v1/agent/chats/chat_123/messages" {
					t.Fatalf("url = %s", req.URL.String())
				}
				if got := req.Header.Get("X-API-Key"); got != "test-key" {
					t.Fatalf("x-api-key = %s", got)
				}
				var payload map[string]any
				if err := json.NewDecoder(req.Body).Decode(&payload); err != nil {
					t.Fatalf("decode request body: %v", err)
				}
				message := payload["message"].(map[string]any)
				if message["content"] != "@alice hello room" {
					t.Fatalf("content = %#v", message["content"])
				}
				if _, ok := message["message_type"]; ok {
					t.Fatalf("message_type should not be sent: %#v", message)
				}
				if _, ok := message["metadata"]; ok {
					t.Fatalf("metadata should not be sent: %#v", message)
				}
			},
			assertResult: func(t *testing.T, out *client.MessageSentResponse, err error) {
				t.Helper()
				if err != nil {
					t.Fatalf("SendChatMessage returned error: %v", err)
				}
				if out == nil || out.ID != "msg_1" || !out.Success {
					t.Fatalf("response = %#v", out)
				}
				if len(out.Recipients) != 1 || out.Recipients[0].Handle != "alice" {
					t.Fatalf("recipients = %#v", out.Recipients)
				}
			},
			wantCalled: true,
		},
		{
			name:         "escapes chat id in request path",
			chatID:       "a/b",
			input:        client.SendChatMessageInput{Content: "@alice hello", Mentions: []client.Mention{{ID: "user_1"}}},
			responseCode: http.StatusOK,
			responseBody: `{"data":{"id":"msg_2","success":true,"recipients":[]}}`,
			assertRequest: func(t *testing.T, req *http.Request) {
				if req.URL.String() != "https://api.test/api/v1/agent/chats/a%2Fb/messages" {
					t.Fatalf("url = %s", req.URL.String())
				}
			},
			assertResult: func(t *testing.T, out *client.MessageSentResponse, err error) {
				if err != nil || out == nil {
					t.Fatalf("out=%#v err=%v", out, err)
				}
			},
			wantCalled: true,
		},
		{
			name:   "returns validation error for missing mentions without network",
			chatID: "chat_123",
			input:  client.SendChatMessageInput{Content: "hello"},
			assertResult: func(t *testing.T, out *client.MessageSentResponse, err error) {
				if out != nil || err == nil || err.Error() != "band: at least one mention is required" {
					t.Fatalf("out=%#v err=%v", out, err)
				}
			},
			wantCalled: false,
		},
		{
			name:   "returns validation error for empty content without network",
			chatID: "chat_123",
			input:  client.SendChatMessageInput{Mentions: []client.Mention{{ID: "user_1"}}},
			assertResult: func(t *testing.T, out *client.MessageSentResponse, err error) {
				if out != nil || err == nil || err.Error() != "band: content is required" {
					t.Fatalf("out=%#v err=%v", out, err)
				}
			},
			wantCalled: false,
		},
		{
			name:         "maps unauthorized response",
			chatID:       "chat_123",
			input:        client.SendChatMessageInput{Content: "@alice hello", Mentions: []client.Mention{{ID: "user_1"}}},
			responseCode: http.StatusUnauthorized,
			responseBody: `{"error":{"code":"unauthorized","message":"bad key"}}`,
			assertResult: func(t *testing.T, out *client.MessageSentResponse, err error) {
				if out != nil || !errors.Is(err, client.ErrUnauthorized) {
					t.Fatalf("out=%#v err=%v", out, err)
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
			sdk := client.New(client.WithApiKey("test-key"), client.WithBaseURL("https://api.test"), client.WithHTTPClient(&http.Client{Transport: transport}))
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

func TestMessageProcessingAndEventEndpoints(t *testing.T) {
	t.Parallel()
	seen := []string{}
	transport := roundTripFunc(func(req *http.Request) (*http.Response, error) {
		seen = append(seen, req.Method+" "+req.URL.String())
		return jsonResponse(http.StatusOK, `{"data":{"id":"msg_1","success":true,"status":"processed","attempt_number":1,"message_type":"task"}}`), nil
	})
	sdk := client.New(client.WithApiKey("test-key"), client.WithBaseURL("https://api.test"), client.WithHTTPClient(&http.Client{Transport: transport}))
	if _, err := sdk.MarkChatMessageProcessing(context.Background(), "chat/1", "msg/1"); err != nil {
		t.Fatal(err)
	}
	if _, err := sdk.MarkChatMessageProcessed(context.Background(), "chat/1", "msg/1"); err != nil {
		t.Fatal(err)
	}
	if _, err := sdk.MarkChatMessageFailed(context.Background(), "chat/1", "msg/1", "boom"); err != nil {
		t.Fatal(err)
	}
	if _, err := sdk.CreateChatEvent(context.Background(), "chat/1", client.CreateChatEventInput{Content: "task", MessageType: "task"}); err != nil {
		t.Fatal(err)
	}
	want := []string{
		"POST https://api.test/api/v1/agent/chats/chat%2F1/messages/msg%2F1/processing",
		"POST https://api.test/api/v1/agent/chats/chat%2F1/messages/msg%2F1/processed",
		"POST https://api.test/api/v1/agent/chats/chat%2F1/messages/msg%2F1/failed",
		"POST https://api.test/api/v1/agent/chats/chat%2F1/events",
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

func TestListChatMessages(t *testing.T) {
	t.Parallel()
	page := 2
	pageSize := 25
	status := "pending"
	transport := roundTripFunc(func(req *http.Request) (*http.Response, error) {
		if req.URL.String() != "https://api.test/api/v1/agent/chats/chat_1/messages?page=2&page_size=25&status=pending" {
			t.Fatalf("url = %s", req.URL.String())
		}
		return jsonResponse(http.StatusOK, `{"data":[{"id":"msg_1","content":"hello","message_type":"text","sender_id":"agent_1","sender_type":"Agent"}],"metadata":{"page":2,"page_size":25}}`), nil
	})
	sdk := client.New(client.WithApiKey("test-key"), client.WithBaseURL("https://api.test"), client.WithHTTPClient(&http.Client{Transport: transport}))
	out, err := sdk.ListChatMessages(context.Background(), "chat_1", &client.ListChatMessagesInput{Page: &page, PageSize: &pageSize, Status: &status})
	if err != nil || out == nil || len(out.Data) != 1 {
		t.Fatalf("out=%#v err=%v", out, err)
	}
}

func TestDeleteChatMessageV2(t *testing.T) {
	t.Parallel()
	transport := roundTripFunc(func(req *http.Request) (*http.Response, error) {
		if req.Method != http.MethodDelete {
			t.Fatalf("method=%s", req.Method)
		}
		if req.URL.String() != "https://api.test/api/v2/chats/chat%2F1/messages/msg%2F1" {
			t.Fatalf("url=%s", req.URL.String())
		}
		return jsonResponse(http.StatusNoContent, ``), nil
	})
	sdk := client.New(client.WithApiKey("test-key"), client.WithBaseURL("https://api.test"), client.WithHTTPClient(&http.Client{Transport: transport}))
	if err := sdk.DeleteChatMessage(context.Background(), "chat/1", "msg/1"); err != nil {
		t.Fatalf("DeleteChatMessage returned error: %v", err)
	}
}

func TestDeleteChatMessageV2ValidationAndAPIErrors(t *testing.T) {
	t.Parallel()
	sdk := client.New(client.WithApiKey("test-key"), client.WithBaseURL("https://api.test"), client.WithHTTPClient(&http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		return jsonResponse(http.StatusUnauthorized, `{"error":{"code":"unauthorized","message":"bad key"}}`), nil
	})}))
	if err := sdk.DeleteChatMessage(context.Background(), "", "msg_1"); err == nil || err.Error() != "band: chat id is required" {
		t.Fatalf("chat validation err=%v", err)
	}
	if err := sdk.DeleteChatMessage(context.Background(), "chat_1", ""); err == nil || err.Error() != "band: message id is required" {
		t.Fatalf("message validation err=%v", err)
	}
	if err := sdk.DeleteChatMessage(context.Background(), "chat_1", "msg_1"); !errors.Is(err, client.ErrUnauthorized) {
		t.Fatalf("DeleteChatMessage err=%v", err)
	}
}
