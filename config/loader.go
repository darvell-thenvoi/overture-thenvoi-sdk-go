package config

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/darvell-thenvoi/overture-thenvoi-sdk-go/core"
	"gopkg.in/yaml.v3"
)

const (
	defaultConfigPath = "./agent_config.yaml"
	defaultEnvPrefix  = "THENVOI_"
)

var camelBoundary = regexp.MustCompile(`([A-Z])`)

var unsafeKeys = map[string]struct{}{
	"__proto__":   {},
	"constructor": {},
	"prototype":   {},
	"toString":    {},
	"to_string":   {},
	"valueOf":     {},
	"value_of":    {},
}

// AgentCredentials contains the credentials needed to authenticate an agent.
type AgentCredentials struct {
	AgentID string
	APIKey  string
	WSURL   string
	RESTURL string
}

// AgentConfigResult contains agent credentials plus safe extra config values.
type AgentConfigResult struct {
	AgentCredentials
	Extra map[string]any
}

// LoadAgentConfigFromEnvOptions configures environment variable loading.
type LoadAgentConfigFromEnvOptions struct {
	Env    map[string]string
	Prefix string
}

// LoadAgentConfig loads agent credentials from a YAML config file.
func LoadAgentConfig(agentKey string, configPath string) (*AgentConfigResult, error) {
	filePath := configPath
	if filePath == "" {
		filePath = defaultConfigPath
	}

	raw, err := os.ReadFile(filePath)
	if err != nil {
		return nil, core.NewValidationError(fmt.Sprintf("Config file not found: %s. Copy agent_config.yaml.example to agent_config.yaml and configure your agents.", filePath))
	}

	var parsed any
	if err := yaml.Unmarshal(raw, &parsed); err != nil {
		return nil, core.NewValidationError(fmt.Sprintf("Invalid config file: %s. Expected a YAML object.", filePath))
	}

	config, ok := toStringMap(parsed)
	if !ok || len(config) == 0 {
		return nil, core.NewValidationError(fmt.Sprintf("Invalid config file: %s. Expected a YAML object.", filePath))
	}

	sourceLabel := filePath
	section := config
	if agentKey != "" {
		keyed, exists := config[agentKey]
		if !exists {
			return nil, core.NewValidationError(fmt.Sprintf("Config key %q not found in %s.", agentKey, filePath))
		}
		keyedMap, ok := toStringMap(keyed)
		if !ok {
			return nil, core.NewValidationError(fmt.Sprintf("Config key %q in %s must be an object with agent_id and api_key.", agentKey, filePath))
		}
		section = keyedMap
		sourceLabel = fmt.Sprintf("%s under key %q", filePath, agentKey)
	}

	return toAgentConfigResult(normalizeKeys(section), sourceLabel)
}

// LoadAgentConfigFromEnv loads agent credentials from environment variables.
func LoadAgentConfigFromEnv(options *LoadAgentConfigFromEnvOptions) (*AgentCredentials, error) {
	env := map[string]string{}
	prefix := defaultEnvPrefix
	if options != nil {
		if options.Env != nil {
			env = options.Env
		} else {
			env = environMap()
		}
		prefix = normalizePrefix(options.Prefix)
	} else {
		env = environMap()
	}

	section := normalizeKeys(map[string]any{
		"agent_id": env[prefix+"AGENT_ID"],
		"api_key":  env[prefix+"API_KEY"],
		"ws_url":   env[prefix+"WS_URL"],
		"rest_url": env[prefix+"REST_URL"],
	})

	result, err := toAgentConfigResult(section, fmt.Sprintf("environment variables (%sAGENT_ID, %sAPI_KEY)", prefix, prefix))
	if err != nil {
		return nil, core.NewValidationError(fmt.Sprintf("%s. Set %sAGENT_ID and %sAPI_KEY, or use loadAgentConfig() for agent_config.yaml.", err.Error(), prefix, prefix))
	}

	credentials := result.AgentCredentials
	return &credentials, nil
}

func toAgentConfigResult(section map[string]any, sourceLabel string) (*AgentConfigResult, error) {
	missing := missingFields(section, "agent_id", "api_key")
	if len(missing) > 0 {
		return nil, core.NewValidationError(fmt.Sprintf("Missing required fields in %s: %s", sourceLabel, strings.Join(missing, ", ")))
	}

	invalid := invalidRequiredFields(section, "agent_id", "api_key")
	if len(invalid) > 0 {
		return nil, core.NewValidationError(fmt.Sprintf("Invalid fields in %s: %s must be non-empty strings", sourceLabel, strings.Join(invalid, ", ")))
	}

	agentID, _ := trimmedString(section["agent_id"])
	apiKey, _ := trimmedString(section["api_key"])

	wsURL, err := optionalString(section, "ws_url", sourceLabel)
	if err != nil {
		return nil, err
	}
	restURL, err := optionalString(section, "rest_url", sourceLabel)
	if err != nil {
		return nil, err
	}

	extra := map[string]any{}
	for key, value := range section {
		if isCoreKey(key) {
			continue
		}
		if _, unsafe := unsafeKeys[key]; unsafe {
			continue
		}
		extra[key] = value
	}

	return &AgentConfigResult{
		AgentCredentials: AgentCredentials{
			AgentID: agentID,
			APIKey:  apiKey,
			WSURL:   wsURL,
			RESTURL: restURL,
		},
		Extra: extra,
	}, nil
}

func missingFields(section map[string]any, fields ...string) []string {
	var missing []string
	for _, field := range fields {
		value, exists := section[field]
		if !exists || value == nil || value == "" {
			missing = append(missing, field)
		}
	}
	return missing
}

func invalidRequiredFields(section map[string]any, fields ...string) []string {
	var invalid []string
	for _, field := range fields {
		value, ok := trimmedString(section[field])
		if !ok || value == "" {
			invalid = append(invalid, field)
		}
	}
	return invalid
}

func optionalString(section map[string]any, field string, sourceLabel string) (string, error) {
	value, exists := section[field]
	if !exists || value == nil {
		return "", nil
	}
	text, ok := value.(string)
	if !ok {
		return "", core.NewValidationError(fmt.Sprintf("%s in %s must be a string", field, sourceLabel))
	}
	return strings.TrimSpace(text), nil
}

func trimmedString(value any) (string, bool) {
	text, ok := value.(string)
	if !ok {
		return "", false
	}
	return strings.TrimSpace(text), true
}

func isCoreKey(key string) bool {
	switch key {
	case "agent_id", "api_key", "ws_url", "rest_url":
		return true
	default:
		return false
	}
}

func normalizeKeys(input map[string]any) map[string]any {
	result := make(map[string]any, len(input))
	for key, value := range input {
		result[toSnakeCase(key)] = value
	}
	return result
}

func toSnakeCase(key string) string {
	return strings.ToLower(camelBoundary.ReplaceAllString(key, "_$1"))
}

func normalizePrefix(prefix string) string {
	if prefix == "" {
		return defaultEnvPrefix
	}
	if strings.HasSuffix(prefix, "_") {
		return prefix
	}
	return prefix + "_"
}

func environMap() map[string]string {
	env := map[string]string{}
	for _, item := range os.Environ() {
		key, value, found := strings.Cut(item, "=")
		if found {
			env[key] = value
		}
	}
	return env
}

func toStringMap(value any) (map[string]any, bool) {
	raw, ok := value.(map[string]any)
	if !ok {
		return nil, false
	}

	result := make(map[string]any, len(raw))
	for key, child := range raw {
		result[key] = child
	}
	return result, true
}
