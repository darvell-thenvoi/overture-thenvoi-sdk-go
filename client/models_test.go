package client_test

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/darvell-thenvoi/overture-thenvoi-sdk-go/client"
)

func TestListModels(t *testing.T) {
	t.Parallel()
	transport := roundTripFunc(func(req *http.Request) (*http.Response, error) {
		if req.Method != http.MethodGet {
			t.Fatalf("method=%s", req.Method)
		}
		if req.URL.String() != "https://api.test/api/v2/models" {
			t.Fatalf("url=%s", req.URL.String())
		}
		return jsonResponse(http.StatusOK, `{"data":[{"id":"model_1","provider":"openai","name":"gpt-4","base_url":"https://api.openai.com"}],"metadata":{"total_models":1,"last_updated":"2026-01-02T03:04:05Z"}}`), nil
	})
	sdk := client.New(client.WithApiKey("test-key"), client.WithBaseURL("https://api.test"), client.WithHTTPClient(&http.Client{Transport: transport}))
	out, err := sdk.ListModels(context.Background())
	if err != nil || len(out.Data) != 1 || out.Metadata.TotalModels != 1 {
		t.Fatalf("ListModels out=%#v err=%v", out, err)
	}
}

func TestListModelsAPIErrors(t *testing.T) {
	t.Parallel()
	sdk := client.New(client.WithApiKey("test-key"), client.WithBaseURL("https://api.test"), client.WithHTTPClient(&http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		return jsonResponse(http.StatusUnauthorized, `{"error":{"code":"unauthorized","message":"bad key"}}`), nil
	})}))
	if _, err := sdk.ListModels(context.Background()); !errors.Is(err, client.ErrUnauthorized) {
		t.Fatalf("ListModels err=%v", err)
	}
}
