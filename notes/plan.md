# Plan: Agent Config Loading

## Summary
Add a Go `config` package that loads agent credentials from `agent_config.yaml` or environment variables with the same public behavior as the TypeScript SDK config loader. Done means callers can use one typed credential value with `AgentCredentials{AgentID, APIKey, WSURL, RESTURL}`, YAML supports keyed and flat shapes with snake_case or camelCase keys, environment loading preserves the `THENVOI_` default prefix, unsafe extra keys are dropped, and every validation failure returns `core.ValidationError`.

## Codebase Context
This repo currently has one Go module, `github.com/darvell-thenvoi/overture-thenvoi-sdk-go`, declared in `go.mod:1`; it targets Go 1.22 at `go.mod:3`. Existing public SDK surfaces are grouped by package area under `client/`, with exported request and response structs plus methods in the same file, as shown by `client/tools.go:1` and `client/agent_admin.go:1`.

The REST client already has package-local API error types in `client/errors.go:8` and `client/errors.go:16`, but local grep found no existing `core` package and no `ValidationError` symbol. Because this work requires `core.ValidationError`, implementation must either use the merged `core` package if it appears in the engineering worktree or add a minimal `core` package with `type ValidationError struct { Message string }`, `func (err *ValidationError) Error() string`, and a constructor/helper only if that matches repo style.

Client configuration uses exported Go struct fields with JSON tags only where wire encoding is needed; see `client/options.go:13` and `client/types.go:9`. The new config structs are not request bodies, so YAML/env mapping should be handled in loader code rather than by exposing snake_case Go fields.

Local grep gate result: no existing local symbols named `AgentCredentials`, `AgentConfigResult`, `LoadAgentConfig`, `LoadAgentConfigFromEnv`, `LoadAgentConfigFromEnvOptions`, or `ValidationError`.

## Upstream Contract Citations
The work mirrors `/Users/pp/thenvoi/thenvoi-sdk-typescript/packages/sdk/src/config/loader.ts`. `overture_read_upstream_contract` could not read it because this work order has no configured `contractRoots`; the cited lines below were read directly from the local upstream repo.

Public type mirrors:

```text
Go config.AgentCredentials mirrors TS AgentCredentials at loader.ts:5.
Go config.AgentConfigResult mirrors TS AgentConfigResult at loader.ts:12, but uses Extra map[string]any instead of an index signature.
Go config.LoadAgentConfigFromEnvOptions mirrors TS LoadAgentConfigFromEnvOptions at loader.ts:79.
Go config.LoadAgentConfig mirrors TS loadAgentConfig at loader.ts:85.
Go config.LoadAgentConfigFromEnv mirrors TS loadAgentConfigFromEnv at loader.ts:129.
```

```ts
# /Users/pp/thenvoi/thenvoi-sdk-typescript/packages/sdk/src/config/loader.ts:3-20
import { ValidationError } from "../core/errors";

export interface AgentCredentials {
  agentId: string;
  apiKey: string;
  wsUrl?: string;
  restUrl?: string;
}

export interface AgentConfigResult extends AgentCredentials {
  [key: string]: unknown;
}

const DEFAULT_CONFIG_PATH = "./agent_config.yaml";
const DEFAULT_ENV_PREFIX = "THENVOI_";

const REQUIRED_FIELDS = ["agent_id", "api_key"] as const;
const UNSAFE_KEYS = new Set(["__proto__", "constructor", "prototype", "toString", "valueOf"]);
```

```ts
# /Users/pp/thenvoi/thenvoi-sdk-typescript/packages/sdk/src/config/loader.ts:22-77
function toSnakeCase(key: string): string {
  return key.replace(/([A-Z])/g, "_$1").toLowerCase();
}

function normalizeKeys(obj: Record<string, unknown>): Record<string, unknown> {
  const result: Record<string, unknown> = {};
  for (const [key, value] of Object.entries(obj)) {
    result[toSnakeCase(key)] = value;
  }
  return result;
}

function toAgentConfigResult(section: Record<string, unknown>, sourceLabel: string): AgentConfigResult {
  const missing = REQUIRED_FIELDS.filter((field) => !section[field]);
  if (missing.length > 0) {
    throw new ValidationError(
      `Missing required fields in ${sourceLabel}: ${missing.join(", ")}`,
    );
  }

  const invalid = REQUIRED_FIELDS.filter(
    (field) => {
      const value = section[field];
      return typeof value !== "string" || value.trim() === "";
    },
  );
  if (invalid.length > 0) {
    throw new ValidationError(
      `Invalid fields in ${sourceLabel}: ${invalid.join(", ")} must be non-empty strings`,
    );
  }

  const { agent_id, api_key, ws_url, rest_url, ...rest } = section;

  if (ws_url !== undefined && typeof ws_url !== "string") {
    throw new ValidationError(`ws_url in ${sourceLabel} must be a string`);
  }
  if (rest_url !== undefined && typeof rest_url !== "string") {
    throw new ValidationError(`rest_url in ${sourceLabel} must be a string`);
  }

  const safeRest: Record<string, unknown> = {};
  for (const [key, value] of Object.entries(rest)) {
    if (!UNSAFE_KEYS.has(key)) {
      safeRest[key] = value;
    }
  }

  return {
    agentId: (agent_id as string).trim(),
    apiKey: (api_key as string).trim(),
    ...(typeof ws_url === "string" && ws_url.trim() !== "" ? { wsUrl: ws_url.trim() } : {}),
    ...(typeof rest_url === "string" && rest_url.trim() !== "" ? { restUrl: rest_url.trim() } : {}),
    ...safeRest,
  };
}
```

```ts
# /Users/pp/thenvoi/thenvoi-sdk-typescript/packages/sdk/src/config/loader.ts:79-156
export interface LoadAgentConfigFromEnvOptions {
  env?: Record<string, string | undefined>;
  prefix?: string;
}

/** Load agent credentials from a YAML config file (defaults to `./agent_config.yaml`). */
export function loadAgentConfig(
  agentKey?: string,
  configPath?: string,
): AgentConfigResult {
  const filePath = configPath ?? DEFAULT_CONFIG_PATH;

  let raw: string;
  try {
    raw = readFileSync(filePath, "utf-8");
  } catch {
    throw new ValidationError(
      `Config file not found: ${filePath}. Copy agent_config.yaml.example to agent_config.yaml and configure your agents.`,
    );
  }

  const parsed = yaml.load(raw, { schema: yaml.JSON_SCHEMA });
  if (!parsed || typeof parsed !== "object") {
    throw new ValidationError(`Invalid config file: ${filePath}. Expected a YAML object.`);
  }

  const config = parsed as Record<string, unknown>;

  // Try keyed format: config[agentKey] is an object with agent_id, api_key
  let section: Record<string, unknown>;
  if (agentKey && agentKey in config) {
    const keyed = config[agentKey];
    if (!keyed || typeof keyed !== "object") {
      throw new ValidationError(
        `Config key "${agentKey}" in ${filePath} must be an object with agent_id and api_key.`,
      );
    }
    section = normalizeKeys(keyed as Record<string, unknown>);
  } else {
    // Flat format: top-level agent_id, api_key
    section = normalizeKeys(config);
  }

  const sourceLabel = agentKey && agentKey in config
    ? `${filePath} under key "${agentKey}"`
    : filePath;
  return toAgentConfigResult(section, sourceLabel);
}

/** Load agent credentials from environment variables (prefix defaults to `THENVOI_`). */
export function loadAgentConfigFromEnv(
  options?: LoadAgentConfigFromEnvOptions,
): AgentCredentials {
  const env = options?.env ?? process.env;
  const prefix = options?.prefix === undefined
    ? DEFAULT_ENV_PREFIX
    : options.prefix === "" || options.prefix.endsWith("_")
      ? options.prefix
      : `${options.prefix}_`;

  const section = normalizeKeys({
    agent_id: env[`${prefix}AGENT_ID`],
    api_key: env[`${prefix}API_KEY`],
    ws_url: env[`${prefix}WS_URL`],
    rest_url: env[`${prefix}REST_URL`],
  });

  try {
    return toAgentConfigResult(section, `environment variables (${prefix}AGENT_ID, ${prefix}API_KEY)`);
  } catch (error) {
    if (error instanceof ValidationError) {
      throw new ValidationError(
        `${error.message}. Set ${prefix}AGENT_ID and ${prefix}API_KEY, or use loadAgentConfig() for agent_config.yaml.`,
      );
    }
    throw error;
  }
}
```

The TS test contract covers keyed YAML, flat YAML, fallback to flat when a key is absent, missing files, missing required fields, camelCase normalization, optional URL mapping, default `THENVOI_` env loading, custom prefix normalization, and env error messages naming required variables; see `/Users/pp/thenvoi/thenvoi-sdk-typescript/packages/sdk/tests/config-loader.test.ts:29`, `:45`, `:57`, `:69`, `:81`, `:91`, `:107`, `:148`, `:164`, and `:177`.

## Goals / Non-Goals / Constraints / Risks
Goals: implement `config.AgentCredentials`, `config.AgentConfigResult`, `config.LoadAgentConfig`, `config.LoadAgentConfigFromEnv`, and `config.LoadAgentConfigFromEnvOptions`; support YAML keyed and flat forms; normalize snake_case and camelCase config keys; trim required and optional string values; omit empty optional URL values; preserve non-core safe extra keys in `Extra`; drop `__proto__`, `constructor`, `prototype`, `toString`, and `valueOf`; return `core.ValidationError` for missing files, malformed YAML, missing required values, invalid optional value types, and missing keyed sections.

Non-goals: no `Agent.Create`, no websocket or REST connection defaults, no runtime constructor changes, no `BAND_` env alias, and no config reads outside this package.

Constraints: empty `configPath` must mean `./agent_config.yaml`; empty `Prefix` must mean `THENVOI_`; a non-empty prefix without trailing underscore must gain one; `THENVOI_` must stay as the default during the rename; `gopkg.in/yaml.v3` must be added to `go.mod` and `go.sum`.

Risks: the ticket says missing keyed sections must be validation errors, while current TS code falls back to flat format when an `agentKey` is not present. Implement the ticket's stricter behavior for Go unless the YAML has a valid flat top-level credential shape and tests require fallback; validation should pin whichever behavior engineering chooses. The absent local `core` package is another risk because the public error contract depends on it.

## Files / Surfaces Expected To Change
- `config/loader.go` — new config package with public structs, env prefix normalization, YAML loading, key normalization, extra filtering, and validation helpers.
- `config/loader_test.go` — new unit tests for keyed and flat YAML, missing and invalid fields, env prefix behavior, optional URLs, and unsafe extra key filtering.
- `core/errors.go` — add only if the merged dependency is absent in this sandbox; needed to expose `core.ValidationError`.
- `agent_config.yaml.example` — new sample keyed config matching the Go package names and required credentials.
- `go.mod` — add `gopkg.in/yaml.v3`.
- `go.sum` — record module checksums for `gopkg.in/yaml.v3`.

## Implementation Approach
Create `package config` with `AgentCredentials` using Go field names `AgentID`, `APIKey`, `WSURL`, and `RESTURL`. Define `AgentConfigResult` as an embedded `AgentCredentials` plus `Extra map[string]any`. Define `LoadAgentConfigFromEnvOptions` with `Env map[string]string` and `Prefix string`.

Use `os.ReadFile` for YAML. If `configPath == ""`, read `./agent_config.yaml`. Decode YAML into `map[string]any` with `yaml.v3`, reject non-map roots, and convert YAML map keys to strings. Normalize keys by converting camelCase to snake_case before validation. For keyed YAML, if `agentKey != ""`, require that key to exist and point to a map; for flat YAML, use the top-level map. Convert required fields only after checking that `agent_id` and `api_key` exist as non-empty strings after trimming. Treat `ws_url` and `rest_url` as optional strings, reject non-string values, and omit them from `AgentCredentials` when absent or blank.

When building `Extra`, exclude `agent_id`, `api_key`, `ws_url`, and `rest_url`, then drop the unsafe keys `__proto__`, `constructor`, `prototype`, `toString`, and `valueOf`. Keep normalized extra key names so camelCase extras do not create duplicate spellings.

For env loading, use the provided `options.Env` when non-nil; otherwise read from `os.Environ` into a map. Normalize `Prefix` with this rule: `""` becomes `THENVOI_`, a non-empty string ending in `_` stays unchanged, and any other non-empty string appends `_`. Read `${prefix}AGENT_ID`, `${prefix}API_KEY`, `${prefix}WS_URL`, and `${prefix}REST_URL`. If validation fails, wrap or create a `core.ValidationError` whose message names `${prefix}AGENT_ID` and `${prefix}API_KEY` and mentions `loadAgentConfig()`.

Tests should use `t.TempDir()` and table-style assertions sparingly. They should check the public result values, `Extra` contents, `errors.As(err, *core.ValidationError)`, and error strings where the ticket requires named env vars.

## Verification Strategy
Validation will run the full Go test suite, build, and vet to catch compile or package-boundary errors. The grep commands prove that the public config symbols required by the architect exist under `config`. The focused loader tests prove YAML shape support, env prefix rules, optional URL handling, unsafe key filtering, and `core.ValidationError` behavior.

## verificationCommands
- `go build ./...`
- `go vet ./...`
- `go test ./...`
- `grep -R "type AgentCredentials" -n config`
- `grep -R "func LoadAgentConfig" -n config`
- `grep -R "func LoadAgentConfigFromEnv" -n config`

## Rollback / Safety Notes
Rollback is limited to removing the new `config` package, optional `core` error file, sample YAML, and the YAML dependency from `go.mod` and `go.sum`. The loader is file/env only and does not change existing client behavior.

## Open Questions
None blocking. The one implementation ambiguity is whether `LoadAgentConfig("missing", path)` should always error for a missing keyed section as the ticket says, or fall back to flat credentials as the current TS code does; the plan chooses the ticket's stricter behavior.
