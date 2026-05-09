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

	tests := []struct {
		name          string
		apiKey        string
		responseCode  int
		responseBody  string
		params        *client.ListChatRoomsParams
		assertRequest func(*testing.T, *http.Request)
		assertResult  func(*testing.T, *client.ListChatRoomsResponse, error)
		wantCalled    bool
	}{
		{
			name:         "decodes success and sends expected request",
			apiKey:       "test-key",
			responseCode: http.StatusOK,
			responseBody: `{"data":[{"id":"chat_1","inserted_at":"2026-01-02T03:04:05Z","task_id":"task_1","title":"General","updated_at":"2026-01-02T03:05:05Z"},{"id":"chat_2","inserted_at":"2026-01-03T03:04:05Z","task_id":null,"title":null,"updated_at":null}],"metadata":{"page":1,"page_size":2,"total_count":2,"total_pages":1}}`,
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
			},
			assertResult: func(t *testing.T, out *client.ListChatRoomsResponse, err error) {
				t.Helper()
				if err != nil {
					t.Fatalf("ListChatRooms returned error: %v", err)
				}
				if out == nil {
					t.Fatal("response is nil")
				}
				if len(out.Data) != 2 {
					t.Fatalf("len(out.Data) = %d", len(out.Data))
				}
				if out.Data[0].ID != "chat_1" {
					t.Fatalf("out[0].id = %s", out.Data[0].ID)
				}
				if out.Data[0].TaskID == nil || *out.Data[0].TaskID != "task_1" {
					t.Fatalf("out[0].task_id = %#v", out.Data[0].TaskID)
				}
				if out.Data[0].Title == nil || *out.Data[0].Title != "General" {
					t.Fatalf("out[0].title = %#v", out.Data[0].Title)
				}
				if out.Data[0].InsertedAt.Format(time.RFC3339) != "2026-01-02T03:04:05Z" {
					t.Fatalf("out[0].inserted_at = %s", out.Data[0].InsertedAt.Format(time.RFC3339))
				}
				if out.Data[0].UpdatedAt == nil || out.Data[0].UpdatedAt.Format(time.RFC3339) != "2026-01-02T03:05:05Z" {
					t.Fatalf("out[0].updated_at = %#v", out.Data[0].UpdatedAt)
				}
				if out.Data[1].TaskID != nil {
					t.Fatalf("out[1].task_id = %#v, want nil", out.Data[1].TaskID)
				}
				if out.Data[1].Title != nil {
					t.Fatalf("out[1].title = %#v, want nil", out.Data[1].Title)
				}
				if out.Data[1].UpdatedAt != nil {
					t.Fatalf("out[1].updated_at = %#v, want nil", out.Data[1].UpdatedAt)
				}
				if out.Metadata.Page != 1 || out.Metadata.PageSize != 2 || out.Metadata.TotalCount != 2 || out.Metadata.TotalPages != 1 {
					t.Fatalf("metadata = %#v", out.Metadata)
				}
			},
			wantCalled: true,
		},
		{
			name:         "encodes pagination query params",
			apiKey:       "test-key",
			responseCode: http.StatusOK,
			params: &client.ListChatRoomsParams{
				Page:     intPtr(2),
				PageSize: intPtr(50),
			},
			responseBody: `{"data":[],"metadata":{"page":2,"page_size":50,"total_count":0,"total_pages":0}}`,
			assertRequest: func(t *testing.T, req *http.Request) {
				t.Helper()
				if req.URL.String() != "https://api.test/api/v1/agent/chats?page=2&page_size=50" {
					t.Fatalf("url = %s", req.URL.String())
				}
			},
			assertResult: func(t *testing.T, out *client.ListChatRoomsResponse, err error) {
				t.Helper()
				if err != nil {
					t.Fatalf("ListChatRooms returned error: %v", err)
				}
				if out == nil {
					t.Fatal("response is nil")
				}
				if out.Metadata.Page != 2 || out.Metadata.PageSize != 50 {
					t.Fatalf("metadata = %#v", out.Metadata)
				}
			},
			wantCalled: true,
		},
		{
			name:         "returns decode error for malformed success body",
			apiKey:       "test-key",
			responseCode: http.StatusOK,
			responseBody: `{"data":`,
			assertResult: func(t *testing.T, out *client.ListChatRoomsResponse, err error) {
				t.Helper()
				if out != nil {
					t.Fatalf("rooms = %#v, want nil", out)
				}
				if err == nil {
					t.Fatal("expected error, got nil")
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
					t.Fatalf("rooms = %#v, want nil", out)
				}
				if err == nil {
					t.Fatal("expected error, got nil")
				}
			},
			wantCalled: false,
		},
		{
			name:         "maps unauthorized response",
			apiKey:       "test-key",
			responseCode: http.StatusUnauthorized,
			responseBody: `{"error":{"code":"unauthorized","message":"bad key"}}`,
			assertResult: func(t *testing.T, out *client.ListChatRoomsResponse, err error) {
				t.Helper()
				if out != nil {
					t.Fatalf("rooms = %#v, want nil", out)
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

			out, err := sdk.ListChatRooms(context.Background(), tt.params)
			tt.assertResult(t, out, err)

			if called != tt.wantCalled {
				t.Fatalf("transport called = %t, want %t", called, tt.wantCalled)
			}
		})
	}
}

func intPtr(v int) *int {
	return &v
}
