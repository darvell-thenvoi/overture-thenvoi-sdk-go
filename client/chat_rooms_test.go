package client_test

import (
	"context"
	"errors"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/darvell-thenvoi/overture-thenvoi-sdk-go/client"
)

func TestListChatRooms(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		apiKey        string
		responseCode  int
		responseBody  string
		assertRequest func(*testing.T, *http.Request)
		assertResult  func(*testing.T, *client.ListChatRoomsResponse, error)
		wantCalled    bool
	}{
		{
			name:         "sends expected request and decodes payload",
			apiKey:       "test-key",
			responseCode: http.StatusOK,
			responseBody: `{"data":[{"id":"chat_1","inserted_at":"2026-01-02T03:04:05Z","updated_at":"2026-01-03T03:04:05Z","task_id":"task_1","title":"Roadmap"},{"id":"chat_2","inserted_at":"2026-01-04T03:04:05Z","updated_at":"2026-01-05T03:04:05Z"}],"metadata":{"page":1,"page_size":50,"total_count":2,"total_pages":1}}`,
			assertRequest: func(t *testing.T, req *http.Request) {
				t.Helper()
				if req.Method != http.MethodGet {
					t.Fatalf("method = %s, want %s", req.Method, http.MethodGet)
				}
				if req.URL.String() != "https://api.test/api/v1/agent/chats" {
					t.Fatalf("url = %s", req.URL.String())
				}
				if got := req.Header.Get("Authorization"); got != "Bearer test-key" {
					t.Fatalf("authorization = %s", got)
				}
				if req.Body == nil {
					return
				}
				body, err := io.ReadAll(req.Body)
				if err != nil {
					t.Fatalf("read request body: %v", err)
				}
				if len(body) != 0 {
					t.Fatalf("request body = %q, want empty", string(body))
				}
			},
			assertResult: func(t *testing.T, out *client.ListChatRoomsResponse, err error) {
				t.Helper()
				if err != nil {
					t.Fatalf("ListChatRooms returned error: %v", err)
				}
				if out == nil {
					t.Fatal("ListChatRooms returned nil response")
				}
				if len(out.Data) != 2 {
					t.Fatalf("len(data) = %d, want 2", len(out.Data))
				}
				if out.Data[0].ID != "chat_1" {
					t.Fatalf("data[0].id = %s", out.Data[0].ID)
				}
				wantInserted, _ := time.Parse(time.RFC3339, "2026-01-02T03:04:05Z")
				if !out.Data[0].InsertedAt.Equal(wantInserted) {
					t.Fatalf("data[0].inserted_at = %s", out.Data[0].InsertedAt)
				}
				wantUpdated, _ := time.Parse(time.RFC3339, "2026-01-03T03:04:05Z")
				if !out.Data[0].UpdatedAt.Equal(wantUpdated) {
					t.Fatalf("data[0].updated_at = %s", out.Data[0].UpdatedAt)
				}
				if out.Data[0].TaskID == nil || *out.Data[0].TaskID != "task_1" {
					t.Fatalf("data[0].task_id = %#v", out.Data[0].TaskID)
				}
				if out.Data[0].Title == nil || *out.Data[0].Title != "Roadmap" {
					t.Fatalf("data[0].title = %#v", out.Data[0].Title)
				}
				if out.Data[1].TaskID != nil {
					t.Fatalf("data[1].task_id = %#v, want nil", out.Data[1].TaskID)
				}
				if out.Data[1].Title != nil {
					t.Fatalf("data[1].title = %#v, want nil", out.Data[1].Title)
				}
				if out.Metadata.Page == nil || *out.Metadata.Page != 1 {
					t.Fatalf("metadata.page = %#v", out.Metadata.Page)
				}
				if out.Metadata.PageSize == nil || *out.Metadata.PageSize != 50 {
					t.Fatalf("metadata.page_size = %#v", out.Metadata.PageSize)
				}
				if out.Metadata.TotalCount == nil || *out.Metadata.TotalCount != 2 {
					t.Fatalf("metadata.total_count = %#v", out.Metadata.TotalCount)
				}
				if out.Metadata.TotalPages == nil || *out.Metadata.TotalPages != 1 {
					t.Fatalf("metadata.total_pages = %#v", out.Metadata.TotalPages)
				}
			},
			wantCalled: true,
		},
		{
			name:         "decodes null optional fields as nil",
			apiKey:       "test-key",
			responseCode: http.StatusOK,
			responseBody: `{"data":[{"id":"chat_1","inserted_at":"2026-01-02T03:04:05Z","updated_at":"2026-01-03T03:04:05Z","task_id":null,"title":null}],"metadata":{}}`,
			assertResult: func(t *testing.T, out *client.ListChatRoomsResponse, err error) {
				t.Helper()
				if err != nil {
					t.Fatalf("ListChatRooms returned error: %v", err)
				}
				if out == nil || len(out.Data) != 1 {
					t.Fatalf("response = %#v", out)
				}
				if out.Data[0].TaskID != nil {
					t.Fatalf("task_id = %#v, want nil", out.Data[0].TaskID)
				}
				if out.Data[0].Title != nil {
					t.Fatalf("title = %#v, want nil", out.Data[0].Title)
				}
			},
			wantCalled: true,
		},
		{
			name:         "maps unauthorized response",
			apiKey:       "test-key",
			responseCode: http.StatusUnauthorized,
			responseBody: `{"error":{"code":"unauthorized","message":"bad key"}}`,
			assertResult: func(t *testing.T, out *client.ListChatRoomsResponse, err error) {
				t.Helper()
				if out != nil {
					t.Fatalf("response = %#v, want nil", out)
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
		{
			name:         "returns decode error for malformed timestamp",
			apiKey:       "test-key",
			responseCode: http.StatusOK,
			responseBody: `{"data":[{"id":"chat_1","inserted_at":"not-a-timestamp","updated_at":"2026-01-03T03:04:05Z"}],"metadata":{}}`,
			assertResult: func(t *testing.T, out *client.ListChatRoomsResponse, err error) {
				t.Helper()
				if out != nil {
					t.Fatalf("response = %#v, want nil", out)
				}
				if err == nil {
					t.Fatal("expected decode error, got nil")
				}
			},
			wantCalled: true,
		},
		{
			name:   "returns missing api key without request",
			apiKey: "",
			assertResult: func(t *testing.T, out *client.ListChatRoomsResponse, err error) {
				t.Helper()
				if out != nil {
					t.Fatalf("response = %#v, want nil", out)
				}
				if err == nil {
					t.Fatal("expected error, got nil")
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

			sdk := client.New(
				client.WithApiKey(tt.apiKey),
				client.WithBaseURL("https://api.test"),
				client.WithHTTPClient(&http.Client{Transport: transport}),
			)

			out, err := sdk.ListChatRooms(context.Background())
			if tt.assertResult != nil {
				tt.assertResult(t, out, err)
			}
			if called != tt.wantCalled {
				t.Fatalf("transport called = %t, want %t", called, tt.wantCalled)
			}
		})
	}
}
