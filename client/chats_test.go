package client_test

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/darvell-thenvoi/overture-thenvoi-sdk-go/client"
)

func TestListChatRooms(t *testing.T) {
	t.Parallel()
	page := 2
	pageSize := 25
	transport := roundTripFunc(func(req *http.Request) (*http.Response, error) {
		if req.Method != http.MethodGet {
			t.Fatalf("method = %s", req.Method)
		}
		if req.URL.String() != "https://api.test/api/v1/agent/chats?page=2&page_size=25" {
			t.Fatalf("url = %s", req.URL.String())
		}
		return jsonResponse(http.StatusOK, `{"data":[{"id":"chat_1","inserted_at":"2026-01-02T03:04:05Z","updated_at":"2026-01-03T03:04:05Z","task_id":"task_1","title":"Roadmap"}],"metadata":{"page":2,"page_size":25,"total_count":1,"total_pages":1}}`), nil
	})
	sdk := client.New(client.WithApiKey("test-key"), client.WithBaseURL("https://api.test"), client.WithHTTPClient(&http.Client{Transport: transport}))
	out, err := sdk.ListChatRooms(context.Background(), &client.ListChatRoomsInput{Page: &page, PageSize: &pageSize})
	if err != nil {
		t.Fatalf("ListChatRooms returned error: %v", err)
	}
	if out == nil || len(out.Data) != 1 {
		t.Fatalf("out=%#v", out)
	}
	if out.Data[0].Title == nil || *out.Data[0].Title != "Roadmap" {
		t.Fatalf("title=%#v", out.Data[0].Title)
	}
	wantInserted, _ := time.Parse(time.RFC3339, "2026-01-02T03:04:05Z")
	if !out.Data[0].InsertedAt.Equal(wantInserted) {
		t.Fatalf("inserted_at=%s", out.Data[0].InsertedAt)
	}
	if out.Metadata.Page == nil || *out.Metadata.Page != 2 {
		t.Fatalf("metadata=%#v", out.Metadata)
	}
}

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
			name:         "decodes contract chat room",
			chatID:       "chat_123",
			responseCode: http.StatusOK,
			responseBody: `{"data":{"id":"chat_123","inserted_at":"2026-01-02T03:04:05Z","updated_at":"2026-01-02T03:06:05Z","task_id":"task_123","title":"Daily Standup"}}`,
			assertRequest: func(t *testing.T, req *http.Request) {
				if req.URL.String() != "https://api.test/api/v1/agent/chats/chat_123" {
					t.Fatalf("url=%s", req.URL.String())
				}
			},
			assertResult: func(t *testing.T, out *client.ChatRoom, err error) {
				if err != nil || out == nil || out.Title == nil || *out.Title != "Daily Standup" {
					t.Fatalf("out=%#v err=%v", out, err)
				}
			},
			wantCalled: true,
		},
		{
			name:         "escapes chat id",
			chatID:       "a/b",
			responseCode: http.StatusOK,
			responseBody: `{"data":{"id":"a/b","inserted_at":"2026-01-02T03:04:05Z","updated_at":"2026-01-02T03:06:05Z"}}`,
			assertRequest: func(t *testing.T, req *http.Request) {
				if req.URL.String() != "https://api.test/api/v1/agent/chats/a%2Fb" {
					t.Fatalf("url=%s", req.URL.String())
				}
			},
			assertResult: func(t *testing.T, out *client.ChatRoom, err error) {
				if err != nil || out == nil {
					t.Fatalf("out=%#v err=%v", out, err)
				}
			},
			wantCalled: true,
		},
		{
			name:   "requires chat id",
			chatID: "",
			assertResult: func(t *testing.T, out *client.ChatRoom, err error) {
				if out != nil || err == nil || err.Error() != "thenvoi: chat id is required" {
					t.Fatalf("out=%#v err=%v", out, err)
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
			out, err := sdk.GetChatRoom(context.Background(), tt.chatID)
			if tt.assertResult != nil {
				tt.assertResult(t, out, err)
			}
			if called != tt.wantCalled {
				t.Fatalf("transport called=%t want %t", called, tt.wantCalled)
			}
		})
	}
}
