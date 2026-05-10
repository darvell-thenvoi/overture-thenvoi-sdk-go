package config

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/darvell-thenvoi/overture-thenvoi-sdk-go/core"
)

func TestLoadAgentConfigKeyedYAML(t *testing.T) {
	path := writeConfig(t, `
supportAgent:
  agentId: " agent-123 "
  apiKey: " api-key "
  wsUrl: " wss://example.test/ws "
  rest_url: " https://example.test "
  displayName: "Support"
  constructor: "dropped"
  toString: "dropped"
`)

	result, err := LoadAgentConfig("supportAgent", path)
	if err != nil {
		t.Fatalf("LoadAgentConfig() error = %v", err)
	}

	if result.AgentID != "agent-123" {
		t.Fatalf("AgentID = %q, want agent-123", result.AgentID)
	}
	if result.APIKey != "api-key" {
		t.Fatalf("APIKey = %q, want api-key", result.APIKey)
	}
	if result.WSURL != "wss://example.test/ws" {
		t.Fatalf("WSURL = %q, want wss://example.test/ws", result.WSURL)
	}
	if result.RESTURL != "https://example.test" {
		t.Fatalf("RESTURL = %q, want https://example.test", result.RESTURL)
	}
	if result.Extra["display_name"] != "Support" {
		t.Fatalf("Extra[display_name] = %v, want Support", result.Extra["display_name"])
	}
	if _, exists := result.Extra["constructor"]; exists {
		t.Fatalf("Extra contains unsafe constructor key")
	}
	if _, exists := result.Extra["to_string"]; exists {
		t.Fatalf("Extra contains unsafe to_string key")
	}
}

func TestLoadAgentConfigFlatYAML(t *testing.T) {
	path := writeConfig(t, `
agent_id: "agent-flat"
api_key: "key-flat"
ws_url: ""
rest_url: "   "
`)

	result, err := LoadAgentConfig("", path)
	if err != nil {
		t.Fatalf("LoadAgentConfig() error = %v", err)
	}

	if result.AgentID != "agent-flat" {
		t.Fatalf("AgentID = %q, want agent-flat", result.AgentID)
	}
	if result.APIKey != "key-flat" {
		t.Fatalf("APIKey = %q, want key-flat", result.APIKey)
	}
	if result.WSURL != "" {
		t.Fatalf("WSURL = %q, want empty", result.WSURL)
	}
	if result.RESTURL != "" {
		t.Fatalf("RESTURL = %q, want empty", result.RESTURL)
	}
}

func TestLoadAgentConfigMissingKeyReturnsValidationError(t *testing.T) {
	path := writeConfig(t, `
agent_id: "agent-flat"
api_key: "key-flat"
`)

	_, err := LoadAgentConfig("missing", path)
	assertValidationError(t, err, `Config key "missing" not found`)
}

func TestLoadAgentConfigValidatesRequiredAndOptionalFields(t *testing.T) {
	missingPath := filepath.Join(t.TempDir(), "missing.yaml")
	_, err := LoadAgentConfig("", missingPath)
	assertValidationError(t, err, "Config file not found")

	invalidRoot := writeConfig(t, `["agent"]`)
	_, err = LoadAgentConfig("", invalidRoot)
	assertValidationError(t, err, "Invalid config file")

	missingRequired := writeConfig(t, `agent_id: "agent"`)
	_, err = LoadAgentConfig("", missingRequired)
	assertValidationError(t, err, "Missing required fields")

	invalidRequired := writeConfig(t, `
agent_id: 123
api_key: "key"
`)
	_, err = LoadAgentConfig("", invalidRequired)
	assertValidationError(t, err, "agent_id must be non-empty strings")

	invalidOptional := writeConfig(t, `
agent_id: "agent"
api_key: "key"
ws_url: 42
`)
	_, err = LoadAgentConfig("", invalidOptional)
	assertValidationError(t, err, "ws_url")
}

func TestLoadAgentConfigFromEnvDefaultPrefix(t *testing.T) {
	result, err := LoadAgentConfigFromEnv(&LoadAgentConfigFromEnvOptions{
		Env: map[string]string{
			"THENVOI_AGENT_ID": " agent-env ",
			"THENVOI_API_KEY":  " key-env ",
			"THENVOI_WS_URL":   " wss://example.test/ws ",
		},
	})
	if err != nil {
		t.Fatalf("LoadAgentConfigFromEnv() error = %v", err)
	}

	if result.AgentID != "agent-env" {
		t.Fatalf("AgentID = %q, want agent-env", result.AgentID)
	}
	if result.APIKey != "key-env" {
		t.Fatalf("APIKey = %q, want key-env", result.APIKey)
	}
	if result.WSURL != "wss://example.test/ws" {
		t.Fatalf("WSURL = %q, want wss://example.test/ws", result.WSURL)
	}
}

func TestLoadAgentConfigFromEnvCustomPrefix(t *testing.T) {
	result, err := LoadAgentConfigFromEnv(&LoadAgentConfigFromEnvOptions{
		Env: map[string]string{
			"BAND_AGENT_ID": "agent-band",
			"BAND_API_KEY":  "key-band",
			"BAND_REST_URL": "https://example.test",
		},
		Prefix: "BAND",
	})
	if err != nil {
		t.Fatalf("LoadAgentConfigFromEnv() error = %v", err)
	}

	if result.AgentID != "agent-band" {
		t.Fatalf("AgentID = %q, want agent-band", result.AgentID)
	}
	if result.RESTURL != "https://example.test" {
		t.Fatalf("RESTURL = %q, want https://example.test", result.RESTURL)
	}
}

func TestLoadAgentConfigFromEnvValidationNamesRequiredVariables(t *testing.T) {
	_, err := LoadAgentConfigFromEnv(&LoadAgentConfigFromEnvOptions{
		Env:    map[string]string{},
		Prefix: "BAND_",
	})

	assertValidationError(t, err, "BAND_AGENT_ID")
	assertErrorContains(t, err, "BAND_API_KEY")
	assertErrorContains(t, err, "loadAgentConfig()")
}

func writeConfig(t *testing.T, contents string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "agent_config.yaml")
	if err := os.WriteFile(path, []byte(strings.TrimSpace(contents)+"\n"), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	return path
}

func assertValidationError(t *testing.T, err error, contains string) {
	t.Helper()

	var validationErr *core.ValidationError
	if !errors.As(err, &validationErr) {
		t.Fatalf("error = %v, want *core.ValidationError", err)
	}
	assertErrorContains(t, err, contains)
}

func assertErrorContains(t *testing.T, err error, contains string) {
	t.Helper()

	if err == nil {
		t.Fatalf("error is nil, want text %q", contains)
	}
	if !strings.Contains(err.Error(), contains) {
		t.Fatalf("error = %q, want text %q", err.Error(), contains)
	}
}
