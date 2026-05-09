package client_test

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/darvell-thenvoi/overture-thenvoi-sdk-go/client"
)

func TestDoDecodesJSONSuccessAndSetsHeaders(t *testing.T) {
	t.Parallel()

	type responseBody struct {
		Name string `json:"name"`
	}
	type requestBody struct {
		AgentID string `json:"agent_id"`
	}

	transport := roundTripFunc(func(r *http.Request) (*http.Response, error) {
		if r.Method != http.MethodPost {
			t.Fatalf("method = %s, want %s", r.Method, http.MethodPost)
		}
		if r.URL.String() != "https://api.test/v1/agents/me?include=profile" {
			t.Fatalf("url = %s", r.URL.String())
		}
		if got := r.Header.Get("Accept"); got != "application/json" {
			t.Fatalf("accept = %s", got)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer test-key" {
			t.Fatalf("authorization = %s", got)
		}
		if got := r.Header.Get("X-API-Key"); got != "test-key" {
			t.Fatalf("x-api-key = %s", got)
		}
		if got := r.Header.Get("User-Agent"); got != "test-agent" {
			t.Fatalf("user-agent = %s", got)
		}
		if got := r.Header.Get("Content-Type"); got != "application/json" {
			t.Fatalf("content-type = %s", got)
		}

		var decoded requestBody
		if err := json.NewDecoder(r.Body).Decode(&decoded); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		if decoded.AgentID != "agent-123" {
			t.Fatalf("agent_id = %s", decoded.AgentID)
		}

		return jsonResponse(http.StatusOK, `{"name":"thenvoi"}`), nil
	})

	sdk := client.New(
		client.WithApiKey("test-key"),
		client.WithBaseURL("https://api.test/"),
		client.WithUserAgent("test-agent"),
		client.WithHTTPClient(&http.Client{Transport: transport}),
	)

	var got responseBody
	err := sdk.Do(
		context.Background(),
		http.MethodPost,
		"/v1/agents/me?include=profile",
		requestBody{AgentID: "agent-123"},
		&got,
	)
	if err != nil {
		t.Fatalf("Do returned error: %v", err)
	}
	if got.Name != "thenvoi" {
		t.Fatalf("name = %s", got.Name)
	}
}

func TestDoDoesNotSetContentTypeWhenBodyIsNil(t *testing.T) {
	t.Parallel()

	transport := roundTripFunc(func(r *http.Request) (*http.Response, error) {
		if got := r.Header.Get("Content-Type"); got != "" {
			t.Fatalf("content-type = %s", got)
		}
		if r.URL.String() != "https://api.test/v1/check" {
			t.Fatalf("url = %s", r.URL.String())
		}
		return jsonResponse(http.StatusOK, `{"ok":true}`), nil
	})

	sdk := client.New(
		client.WithApiKey("test-key"),
		client.WithBaseURL("https://api.test"),
		client.WithHTTPClient(&http.Client{Transport: transport}),
	)

	var got struct {
		OK bool `json:"ok"`
	}
	if err := sdk.Do(context.Background(), http.MethodGet, "v1/check", nil, &got); err != nil {
		t.Fatalf("Do returned error: %v", err)
	}
	if !got.OK {
		t.Fatal("ok = false")
	}
}

func TestDoUsesDefaultBaseURLAndUserAgent(t *testing.T) {
	t.Parallel()

	transport := roundTripFunc(func(r *http.Request) (*http.Response, error) {
		if r.URL.String() != client.DefaultBaseURL+"/v1/check" {
			t.Fatalf("url = %s", r.URL.String())
		}
		if got := r.Header.Get("User-Agent"); got != "overture-thenvoi-sdk-go/"+client.Version {
			t.Fatalf("user-agent = %s", got)
		}
		return jsonResponse(http.StatusOK, `{"ok":true}`), nil
	})

	sdk := client.New(
		client.WithApiKey("test-key"),
		client.WithHTTPClient(&http.Client{Transport: transport}),
	)

	var got struct {
		OK bool `json:"ok"`
	}
	if err := sdk.Do(context.Background(), http.MethodGet, "/v1/check", nil, &got); err != nil {
		t.Fatalf("Do returned error: %v", err)
	}
	if !got.OK {
		t.Fatal("ok = false")
	}
}

func TestDoMissingApiKeyDoesNotCallNetwork(t *testing.T) {
	t.Parallel()

	called := false
	transport := roundTripFunc(func(r *http.Request) (*http.Response, error) {
		called = true
		return jsonResponse(http.StatusOK, `{}`), nil
	})

	sdk := client.New(
		client.WithBaseURL("https://api.test"),
		client.WithHTTPClient(&http.Client{Transport: transport}),
	)

	err := sdk.Do(context.Background(), http.MethodGet, "/v1/check", nil, nil)
	if err == nil {
		t.Fatal("Do returned nil error")
	}
	if called {
		t.Fatal("server was called without api key")
	}
}

func TestDoHonorsContextCancellation(t *testing.T) {
	t.Parallel()

	transport := roundTripFunc(func(r *http.Request) (*http.Response, error) {
		return nil, r.Context().Err()
	})

	sdk := client.New(
		client.WithApiKey("test-key"),
		client.WithBaseURL("https://api.test"),
		client.WithHTTPClient(&http.Client{Transport: transport}),
	)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := sdk.Do(ctx, http.MethodGet, "/v1/check", nil, nil)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("err = %v, want context.Canceled", err)
	}
}

func TestDoReturnsUnauthorizedApiError(t *testing.T) {
	t.Parallel()

	transport := roundTripFunc(func(r *http.Request) (*http.Response, error) {
		resp := jsonResponse(http.StatusUnauthorized, `{"error":{"code":"unauthorized","message":"bad key"}}`)
		resp.Header.Set("X-Request-ID", "req-401")
		return resp, nil
	})

	sdk := client.New(
		client.WithApiKey("test-key"),
		client.WithBaseURL("https://api.test"),
		client.WithHTTPClient(&http.Client{Transport: transport}),
	)

	err := sdk.Do(context.Background(), http.MethodGet, "/v1/check", nil, nil)
	if !errors.Is(err, client.ErrUnauthorized) {
		t.Fatalf("err = %v, want ErrUnauthorized", err)
	}

	var apiErr *client.ApiError
	if !errors.As(err, &apiErr) {
		t.Fatalf("err = %T, want *client.ApiError", err)
	}
	if apiErr.StatusCode != http.StatusUnauthorized || apiErr.Code != "unauthorized" || apiErr.Message != "bad key" {
		t.Fatalf("api error = %+v", apiErr)
	}
	if apiErr.RequestID != "req-401" {
		t.Fatalf("request id = %s", apiErr.RequestID)
	}
}

func TestDoReturnsRateLimitedApiErrorWithRetryAfter(t *testing.T) {
	t.Parallel()

	transport := roundTripFunc(func(r *http.Request) (*http.Response, error) {
		resp := jsonResponse(http.StatusTooManyRequests, `{"code":"rate_limited","message":"slow down"}`)
		resp.Header.Set("Retry-After", "60")
		return resp, nil
	})

	sdk := client.New(
		client.WithApiKey("test-key"),
		client.WithBaseURL("https://api.test"),
		client.WithHTTPClient(&http.Client{Transport: transport}),
	)

	err := sdk.Do(context.Background(), http.MethodGet, "/v1/check", nil, nil)
	if !errors.Is(err, client.ErrRateLimited) {
		t.Fatalf("err = %v, want ErrRateLimited", err)
	}

	var apiErr *client.ApiError
	if !errors.As(err, &apiErr) {
		t.Fatalf("err = %T, want *client.ApiError", err)
	}
	if apiErr.RetryAfter != "60" {
		t.Fatalf("retry after = %s", apiErr.RetryAfter)
	}
}

func TestDoReturnsServerApiErrorForNonJSONBody(t *testing.T) {
	t.Parallel()

	transport := roundTripFunc(func(r *http.Request) (*http.Response, error) {
		return textResponse(http.StatusInternalServerError, "upstream failed"), nil
	})

	sdk := client.New(
		client.WithApiKey("test-key"),
		client.WithBaseURL("https://api.test"),
		client.WithHTTPClient(&http.Client{Transport: transport}),
	)

	err := sdk.Do(context.Background(), http.MethodGet, "/v1/check", nil, nil)
	if !errors.Is(err, client.ErrServer) {
		t.Fatalf("err = %v, want ErrServer", err)
	}

	var apiErr *client.ApiError
	if !errors.As(err, &apiErr) {
		t.Fatalf("err = %T, want *client.ApiError", err)
	}
	if apiErr.Body != "upstream failed" {
		t.Fatalf("body = %q", apiErr.Body)
	}
	if got := apiErr.Error(); got != http.StatusText(http.StatusInternalServerError) {
		t.Fatalf("Error() = %s", got)
	}
}

func TestNewUsesDefaultsAndDoesNotMutateCustomHTTPClient(t *testing.T) {
	t.Parallel()

	customHTTPClient := &http.Client{Timeout: time.Second}
	_ = client.New(client.WithApiKey("test-key"), client.WithHTTPClient(customHTTPClient))

	if customHTTPClient.Timeout != time.Second {
		t.Fatalf("custom timeout = %s", customHTTPClient.Timeout)
	}
}

func TestDoRejectsMalformedBaseURLAtRequestTime(t *testing.T) {
	t.Parallel()

	sdk := client.New(client.WithApiKey("test-key"), client.WithBaseURL("://bad-url"))

	err := sdk.Do(context.Background(), http.MethodGet, "/v1/check", nil, nil)
	if err == nil {
		t.Fatal("Do returned nil error")
	}
	if !strings.Contains(err.Error(), "invalid base url") {
		t.Fatalf("err = %v, want invalid base url error", err)
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (fn roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return fn(r)
}

func jsonResponse(statusCode int, body string) *http.Response {
	resp := textResponse(statusCode, body)
	resp.Header.Set("Content-Type", "application/json")
	return resp
}

func textResponse(statusCode int, body string) *http.Response {
	return &http.Response{
		StatusCode: statusCode,
		Status:     http.StatusText(statusCode),
		Header:     make(http.Header),
		Body:       io.NopCloser(strings.NewReader(body)),
	}
}
