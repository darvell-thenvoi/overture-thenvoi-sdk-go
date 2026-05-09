package client_test

import (
	"context"
	"errors"
	"net/http"
	"testing"

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
		assertResult  func(*testing.T, []client.ChatRoom, error)
		wantCalled    bool
	}{
		{
			name:         "decodes success and sends expected request",
			apiKey:       "test-key",
			responseCode: http.StatusOK,
			responseBody: `{"data":[{"id":"chat_1","name":"General"},{"id":"chat_2","name":null}]}`,
			assertRequest: func(t *testing.T, req *http.Request) {
				t.Helper()
				if req.Method != http.MethodGet {
					t.Fatalf("method = %s, want %s", req.Method, http.MethodGet)
				}
				if req.URL.String() != "https://api.test/v1/chats" {
					t.Fatalf("url = %s", req.URL.String())
				}
				if got := req.Header.Get("Authorization"); got != "Bearer test-key" {
					t.Fatalf("authorization = %s", got)
				}
			},
			assertResult: func(t *testing.T, out []client.ChatRoom, err error) {
				t.Helper()
				if err != nil {
					t.Fatalf("ListChatRooms returned error: %v", err)
				}
				if len(out) != 2 {
					t.Fatalf("len(out) = %d", len(out))
				}
				if out[0].ID != "chat_1" {
					t.Fatalf("out[0].id = %s", out[0].ID)
				}
				if out[0].Name == nil || *out[0].Name != "General" {
					t.Fatalf("out[0].name = %#v", out[0].Name)
				}
				if out[1].Name != nil {
					t.Fatalf("out[1].name = %#v, want nil", out[1].Name)
				}
			},
			wantCalled: true,
		},
		{
			name:         "returns decode error for malformed success body",
			apiKey:       "test-key",
			responseCode: http.StatusOK,
			responseBody: `{"data":`,
			assertResult: func(t *testing.T, out []client.ChatRoom, err error) {
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
			assertResult: func(t *testing.T, out []client.ChatRoom, err error) {
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
			assertResult: func(t *testing.T, out []client.ChatRoom, err error) {
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

			out, err := sdk.ListChatRooms(context.Background())
			tt.assertResult(t, out, err)

			if called != tt.wantCalled {
				t.Fatalf("transport called = %t, want %t", called, tt.wantCalled)
			}
		})
	}
}
