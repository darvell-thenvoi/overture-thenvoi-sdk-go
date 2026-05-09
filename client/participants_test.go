package client_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"testing"

	"github.com/darvell-thenvoi/overture-thenvoi-sdk-go/client"
)

func TestListChatParticipants(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		chatID        string
		responseCode  int
		responseBody  string
		assertRequest func(*testing.T, *http.Request)
		assertResult  func(*testing.T, []client.ChatParticipant, error)
		wantCalled    bool
	}{
		{
			name:         "decodes response and sends expected request",
			chatID:       "chat_123",
			responseCode: http.StatusOK,
			responseBody: `{"data":[{"id":"user_1","handle":"@alice","name":"Alice","role":"owner","status":"active","type":"User"},{"id":"agent_1","status":"active","role":"member","type":"Agent"}]}`,
			assertRequest: func(t *testing.T, req *http.Request) {
				t.Helper()
				if req.Method != http.MethodGet {
					t.Fatalf("method = %s, want %s", req.Method, http.MethodGet)
				}
				if req.URL.String() != "https://api.test/api/v1/agent/chats/chat_123/participants" {
					t.Fatalf("url = %s", req.URL.String())
				}
				if got := req.Header.Get("Authorization"); got != "Bearer test-key" {
					t.Fatalf("authorization = %s", got)
				}
			},
			assertResult: func(t *testing.T, out []client.ChatParticipant, err error) {
				t.Helper()
				if err != nil {
					t.Fatalf("ListChatParticipants returned error: %v", err)
				}
				if len(out) != 2 {
					t.Fatalf("len = %d", len(out))
				}
				if out[0].ID != "user_1" || out[0].Role != client.ParticipantRoleOwner || out[0].Type != client.ChatParticipantTypeUser {
					t.Fatalf("participant[0] = %#v", out[0])
				}
				if out[0].Handle == nil || *out[0].Handle != "@alice" {
					t.Fatalf("handle = %#v", out[0].Handle)
				}
				if out[1].Name != nil {
					t.Fatalf("name = %#v, want nil", out[1].Name)
				}
			},
			wantCalled: true,
		},
		{
			name:         "escapes chat id in request path",
			chatID:       "chat/a",
			responseCode: http.StatusOK,
			responseBody: `{"data":[]}`,
			assertRequest: func(t *testing.T, req *http.Request) {
				t.Helper()
				if req.URL.String() != "https://api.test/api/v1/agent/chats/chat%2Fa/participants" {
					t.Fatalf("url = %s", req.URL.String())
				}
			},
			assertResult: func(t *testing.T, out []client.ChatParticipant, err error) {
				t.Helper()
				if err != nil {
					t.Fatalf("ListChatParticipants returned error: %v", err)
				}
			},
			wantCalled: true,
		},
		{
			name:   "returns validation error for empty chat id without network",
			chatID: "",
			assertResult: func(t *testing.T, out []client.ChatParticipant, err error) {
				t.Helper()
				if out != nil {
					t.Fatalf("participants = %#v, want nil", out)
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
			assertResult: func(t *testing.T, out []client.ChatParticipant, err error) {
				t.Helper()
				if out != nil {
					t.Fatalf("participants = %#v, want nil", out)
				}
				if !errors.Is(err, client.ErrUnauthorized) {
					t.Fatalf("err = %v, want ErrUnauthorized", err)
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

			out, err := sdk.ListChatParticipants(context.Background(), tt.chatID)
			if tt.assertResult != nil {
				tt.assertResult(t, out, err)
			}

			if called != tt.wantCalled {
				t.Fatalf("called = %v, want %v", called, tt.wantCalled)
			}
		})
	}
}

func TestAddChatParticipant(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		chatID        string
		input         client.AddChatParticipantInput
		responseCode  int
		responseBody  string
		assertRequest func(*testing.T, *http.Request)
		assertResult  func(*testing.T, *client.ChatParticipant, error)
		wantCalled    bool
	}{
		{
			name:   "sends expected payload and decodes response",
			chatID: "chat_123",
			input: client.AddChatParticipantInput{
				ParticipantID: "agent_9",
				Role:          client.ParticipantRoleAdmin,
			},
			responseCode: http.StatusOK,
			responseBody: `{"data":{"id":"agent_9","role":"admin","status":"active","type":"Agent","handle":"@team/bot"}}`,
			assertRequest: func(t *testing.T, req *http.Request) {
				t.Helper()
				if req.Method != http.MethodPost {
					t.Fatalf("method = %s, want %s", req.Method, http.MethodPost)
				}
				if req.URL.String() != "https://api.test/api/v1/agent/chats/chat_123/participants" {
					t.Fatalf("url = %s", req.URL.String())
				}
				if got := req.Header.Get("Content-Type"); got != "application/json" {
					t.Fatalf("content-type = %s", got)
				}

				var payload map[string]any
				if err := json.NewDecoder(req.Body).Decode(&payload); err != nil {
					t.Fatalf("decode request body: %v", err)
				}
				participant, ok := payload["participant"].(map[string]any)
				if !ok {
					t.Fatalf("participant wrapper missing: %#v", payload)
				}
				if participant["participant_id"] != "agent_9" {
					t.Fatalf("participant_id = %#v", participant["participant_id"])
				}
				if participant["role"] != "admin" {
					t.Fatalf("role = %#v", participant["role"])
				}
			},
			assertResult: func(t *testing.T, out *client.ChatParticipant, err error) {
				t.Helper()
				if err != nil {
					t.Fatalf("AddChatParticipant returned error: %v", err)
				}
				if out == nil {
					t.Fatal("AddChatParticipant returned nil participant")
				}
				if out.ID != "agent_9" || out.Role != client.ParticipantRoleAdmin {
					t.Fatalf("participant = %#v", out)
				}
			},
			wantCalled: true,
		},
		{
			name:   "omits empty role in payload",
			chatID: "chat_123",
			input: client.AddChatParticipantInput{
				ParticipantID: "user_2",
			},
			responseCode: http.StatusOK,
			responseBody: `{"data":{"id":"user_2","role":"member","status":"active","type":"User"}}`,
			assertRequest: func(t *testing.T, req *http.Request) {
				t.Helper()
				var payload map[string]any
				if err := json.NewDecoder(req.Body).Decode(&payload); err != nil {
					t.Fatalf("decode request body: %v", err)
				}
				participant, ok := payload["participant"].(map[string]any)
				if !ok {
					t.Fatalf("participant wrapper missing: %#v", payload)
				}
				if _, ok := participant["role"]; ok {
					t.Fatalf("role should be omitted: %#v", participant)
				}
			},
			assertResult: func(t *testing.T, out *client.ChatParticipant, err error) {
				t.Helper()
				if err != nil {
					t.Fatalf("AddChatParticipant returned error: %v", err)
				}
				if out == nil || out.ID != "user_2" {
					t.Fatalf("participant = %#v", out)
				}
			},
			wantCalled: true,
		},
		{
			name:   "escapes chat id in request path",
			chatID: "a/b",
			input: client.AddChatParticipantInput{
				ParticipantID: "user_2",
			},
			responseCode: http.StatusOK,
			responseBody: `{"data":{"id":"user_2","role":"member","status":"active","type":"User"}}`,
			assertRequest: func(t *testing.T, req *http.Request) {
				t.Helper()
				if req.URL.String() != "https://api.test/api/v1/agent/chats/a%2Fb/participants" {
					t.Fatalf("url = %s", req.URL.String())
				}
			},
			assertResult: func(t *testing.T, out *client.ChatParticipant, err error) {
				t.Helper()
				if err != nil {
					t.Fatalf("AddChatParticipant returned error: %v", err)
				}
			},
			wantCalled: true,
		},
		{
			name:   "returns validation error for empty chat id without network",
			chatID: "",
			input: client.AddChatParticipantInput{
				ParticipantID: "user_2",
			},
			assertResult: func(t *testing.T, out *client.ChatParticipant, err error) {
				t.Helper()
				if out != nil {
					t.Fatalf("participant = %#v, want nil", out)
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
			name:   "returns validation error for empty participant id without network",
			chatID: "chat_123",
			input:  client.AddChatParticipantInput{},
			assertResult: func(t *testing.T, out *client.ChatParticipant, err error) {
				t.Helper()
				if out != nil {
					t.Fatalf("participant = %#v, want nil", out)
				}
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if err.Error() != "thenvoi: participant id is required" {
					t.Fatalf("err = %v", err)
				}
			},
			wantCalled: false,
		},
		{
			name:   "maps unauthorized response",
			chatID: "chat_123",
			input: client.AddChatParticipantInput{
				ParticipantID: "user_2",
			},
			responseCode: http.StatusUnauthorized,
			responseBody: `{"error":{"code":"unauthorized","message":"bad key"}}`,
			assertResult: func(t *testing.T, out *client.ChatParticipant, err error) {
				t.Helper()
				if out != nil {
					t.Fatalf("participant = %#v, want nil", out)
				}
				if !errors.Is(err, client.ErrUnauthorized) {
					t.Fatalf("err = %v, want ErrUnauthorized", err)
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

			out, err := sdk.AddChatParticipant(context.Background(), tt.chatID, tt.input)
			if tt.assertResult != nil {
				tt.assertResult(t, out, err)
			}

			if called != tt.wantCalled {
				t.Fatalf("called = %v, want %v", called, tt.wantCalled)
			}
		})
	}
}

func TestRemoveChatParticipant(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		chatID        string
		participantID string
		responseCode  int
		responseBody  string
		assertRequest func(*testing.T, *http.Request)
		assertResult  func(*testing.T, *client.ChatParticipant, error)
		wantCalled    bool
	}{
		{
			name:          "sends expected request and decodes response",
			chatID:        "chat_123",
			participantID: "participant_7",
			responseCode:  http.StatusOK,
			responseBody:  `{"data":{"id":"participant_7","role":"member","status":"removed","type":"User"}}`,
			assertRequest: func(t *testing.T, req *http.Request) {
				t.Helper()
				if req.Method != http.MethodDelete {
					t.Fatalf("method = %s, want %s", req.Method, http.MethodDelete)
				}
				if req.URL.String() != "https://api.test/api/v1/agent/chats/chat_123/participants/participant_7" {
					t.Fatalf("url = %s", req.URL.String())
				}
			},
			assertResult: func(t *testing.T, out *client.ChatParticipant, err error) {
				t.Helper()
				if err != nil {
					t.Fatalf("RemoveChatParticipant returned error: %v", err)
				}
				if out == nil || out.ID != "participant_7" {
					t.Fatalf("participant = %#v", out)
				}
			},
			wantCalled: true,
		},
		{
			name:          "escapes chat and participant ids in path",
			chatID:        "chat/a",
			participantID: "participant/b",
			responseCode:  http.StatusOK,
			responseBody:  `{"data":{"id":"participant_7","role":"member","status":"removed","type":"User"}}`,
			assertRequest: func(t *testing.T, req *http.Request) {
				t.Helper()
				if req.URL.String() != "https://api.test/api/v1/agent/chats/chat%2Fa/participants/participant%2Fb" {
					t.Fatalf("url = %s", req.URL.String())
				}
			},
			assertResult: func(t *testing.T, out *client.ChatParticipant, err error) {
				t.Helper()
				if err != nil {
					t.Fatalf("RemoveChatParticipant returned error: %v", err)
				}
			},
			wantCalled: true,
		},
		{
			name:          "returns validation error for empty chat id without network",
			chatID:        "",
			participantID: "participant_7",
			assertResult: func(t *testing.T, out *client.ChatParticipant, err error) {
				t.Helper()
				if out != nil {
					t.Fatalf("participant = %#v, want nil", out)
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
			name:          "returns validation error for empty participant id without network",
			chatID:        "chat_123",
			participantID: "",
			assertResult: func(t *testing.T, out *client.ChatParticipant, err error) {
				t.Helper()
				if out != nil {
					t.Fatalf("participant = %#v, want nil", out)
				}
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if err.Error() != "thenvoi: participant id is required" {
					t.Fatalf("err = %v", err)
				}
			},
			wantCalled: false,
		},
		{
			name:          "maps unauthorized response",
			chatID:        "chat_123",
			participantID: "participant_7",
			responseCode:  http.StatusUnauthorized,
			responseBody:  `{"error":{"code":"unauthorized","message":"bad key"}}`,
			assertResult: func(t *testing.T, out *client.ChatParticipant, err error) {
				t.Helper()
				if out != nil {
					t.Fatalf("participant = %#v, want nil", out)
				}
				if !errors.Is(err, client.ErrUnauthorized) {
					t.Fatalf("err = %v, want ErrUnauthorized", err)
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

			out, err := sdk.RemoveChatParticipant(context.Background(), tt.chatID, tt.participantID)
			if tt.assertResult != nil {
				tt.assertResult(t, out, err)
			}

			if called != tt.wantCalled {
				t.Fatalf("called = %v, want %v", called, tt.wantCalled)
			}
		})
	}
}
