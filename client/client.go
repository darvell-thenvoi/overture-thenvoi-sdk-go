package client

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// Client is the REST client for the Band platform API.
type Client struct {
	config Config
}

// New creates a REST client with defaults and the supplied options.
func New(opts ...Option) *Client {
	config := Config{
		BaseURL:   DefaultBaseURL,
		UserAgent: "band-sdk-go/" + Version,
		Timeout:   defaultTimeout,
	}

	for _, opt := range opts {
		opt(&config)
	}

	if config.HTTPClient == nil {
		base := *http.DefaultClient
		base.Timeout = config.Timeout
		config.HTTPClient = &base
	}

	return &Client{config: config}
}

// Do sends an HTTP request and decodes a successful JSON response into out.
func (client *Client) Do(ctx context.Context, method string, path string, body any, out any) error {
	_, err := client.do(ctx, method, path, body, out)
	return err
}

func (client *Client) do(ctx context.Context, method string, path string, body any, out any) (int, error) {
	if client.config.ApiKey == "" {
		return 0, errors.New("band: api key is required")
	}

	requestURL, err := client.requestURL(path)
	if err != nil {
		return 0, err
	}

	var requestBody io.Reader
	if body != nil {
		var buffer bytes.Buffer
		if err := json.NewEncoder(&buffer).Encode(body); err != nil {
			return 0, err
		}
		requestBody = &buffer
	}

	req, err := http.NewRequestWithContext(ctx, method, requestURL, requestBody)
	if err != nil {
		return 0, err
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-API-Key", client.config.ApiKey)
	if client.config.UserAgent != "" {
		req.Header.Set("User-Agent", client.config.UserAgent)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := client.config.HTTPClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return resp.StatusCode, newAPIError(resp)
	}

	if out == nil || resp.StatusCode == http.StatusNoContent {
		return resp.StatusCode, nil
	}

	return resp.StatusCode, json.NewDecoder(resp.Body).Decode(out)
}

func (client *Client) requestURL(path string) (string, error) {
	baseURL := strings.TrimSuffix(client.config.BaseURL, "/")
	parsedBaseURL, err := url.Parse(baseURL)
	if err != nil {
		return "", fmt.Errorf("band: invalid base url: %w", err)
	}
	if parsedBaseURL.Scheme == "" || parsedBaseURL.Host == "" {
		return "", fmt.Errorf("band: invalid base url %q", client.config.BaseURL)
	}

	if path == "" {
		return parsedBaseURL.String(), nil
	}

	requestPath := path
	if !strings.HasPrefix(requestPath, "/") {
		requestPath = "/" + requestPath
	}
	return parsedBaseURL.String() + requestPath, nil
}

type apiErrorBody struct {
	Code    string       `json:"code"`
	Message string       `json:"message"`
	Error   apiErrorInfo `json:"error"`
}

type apiErrorInfo struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func newAPIError(resp *http.Response) error {
	body, readErr := io.ReadAll(resp.Body)
	apiErr := &ApiError{
		StatusCode: resp.StatusCode,
		RequestID:  resp.Header.Get("X-Request-ID"),
		RetryAfter: resp.Header.Get("Retry-After"),
	}
	if readErr == nil {
		apiErr.Body = string(body)
		parseAPIErrorBody(body, apiErr)
	}
	return apiErr
}

func parseAPIErrorBody(body []byte, apiErr *ApiError) {
	if len(body) == 0 {
		return
	}

	var parsed apiErrorBody
	if err := json.Unmarshal(body, &parsed); err != nil {
		return
	}

	apiErr.Code = parsed.Code
	apiErr.Message = parsed.Message
	if parsed.Error.Code != "" {
		apiErr.Code = parsed.Error.Code
	}
	if parsed.Error.Message != "" {
		apiErr.Message = parsed.Error.Message
	}
}
