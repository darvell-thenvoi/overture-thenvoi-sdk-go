# Plan: Core SDK Contracts, Errors, and Generic Adapter

## Summary
Add the foundational Go SDK packages that later runtime and adapter work orders can import. Done means `core` exposes the DTOs, history provider, protocols, typed errors, logger, and tool executor helper types named in the work order; `adapters` exposes `GenericAdapter` and `GenericAdapterHandler`; tests prove error wrapping, history conversion, adapter lifecycle forwarding, logger no-op behavior, and tool executor error helpers.

## Codebase Context
This repo currently contains only the `client` package plus docs. `go.mod:1` declares module `github.com/darvell-thenvoi/overture-thenvoi-sdk-go`, so new packages should import each other through that module path when cross-package imports are needed.

Existing DTO style uses exported Go structs with JSON tags and pointer fields for optional API values, for example `client/types.go:9` and `client/types.go:17`. Existing tests are package-level Go tests with standard `testing` and `httptest`/fake helpers rather than external assertion libraries, as seen in `client/client_test.go:303` and `client/chats_test.go:17`.

Local symbol grep found no existing `FrameworkAdapter`, `AdapterToolsProtocol`, `ThenvoiSdkError`, `GenericAdapter`, `HistoryProvider`, `ToolOperationResult`, `MentionReference`, `PaginationMetadataLike`, `PaginatedList`, `AgentToolsCapabilities`, or `NoopLogger` definitions in this lease, so the new public symbols will not duplicate a shipped local surface.

## Upstream Contract Citations
The work mirrors the TypeScript SDK public contracts under `/Users/pp/thenvoi/thenvoi-sdk-typescript/packages/sdk/src`. `overture_read_upstream_contract` could not read these files because this work order has no configured contract roots, so citations below come from direct local reads with line numbers.

DTO mapping:

```typescript
// /Users/pp/thenvoi/thenvoi-sdk-typescript/packages/sdk/src/contracts/dtos.ts:1-31
export interface MetadataMap {
  [key: string]: unknown;
}

export interface ToolOperationResult {
  ok?: boolean;
  status?: string;
  [key: string]: unknown;
}

export interface MentionReference {
  id: string;
  handle?: string;
  name?: string;
  username?: string;
}

export type MentionInput = string[] | MentionReference[];

export interface PaginationMetadataLike {
  page?: number;
  pageSize?: number;
  totalPages?: number;
  totalCount?: number;
  [key: string]: unknown;
}

export interface PaginatedList<TItem = MetadataMap> {
  data: TItem[];
  metadata?: PaginationMetadataLike;
}
```

Record DTO mapping:

```typescript
// /Users/pp/thenvoi/thenvoi-sdk-typescript/packages/sdk/src/contracts/dtos.ts:33-59
export interface ParticipantRecord {
  id: string;
  name: string;
  type: string;
  handle?: string | null;
  role?: string;
}

export interface PeerRecord {
  id?: string;
  name?: string;
  type?: string;
  handle?: string | null;
  description?: string | null;
}

// Wire DTOs intentionally preserve API snake_case field names.
export interface WireContactRecord {
  id?: string;
  handle?: string;
  name?: string | null;
  type?: string;
  description?: string | null;
  is_external?: boolean | null;
  inserted_at?: string;
}
export type ContactRecord = WireContactRecord;
```

Contact, memory, and tool schema DTO mapping:

```typescript
// /Users/pp/thenvoi/thenvoi-sdk-typescript/packages/sdk/src/contracts/dtos.ts:81-172
export interface WireContactRequestsResult {
  received: WireReceivedContactRequestRecord[];
  sent: WireSentContactRequestRecord[];
  metadata?: MetadataMap;
}
export type ContactRequestsResult = WireContactRequestsResult;

export type ContactRequestAction = "approve" | "reject" | "cancel";

export interface ListContactsArgs {
  page?: number;
  pageSize?: number;
}

export interface AddContactArgs {
  handle: string;
  message?: string;
}

export type RemoveContactArgs =
  | { target: "handle"; handle: string }
  | { target: "contactId"; contactId: string };

export interface ListContactRequestsArgs {
  page?: number;
  pageSize?: number;
  sentStatus?: string;
}

export type RespondContactRequestArgs =
  | { action: ContactRequestAction; target: "handle"; handle: string }
  | { action: ContactRequestAction; target: "requestId"; requestId: string };

export interface WireMemoryRecord {
  id?: string;
  content?: string;
  system?: string;
  type?: string;
  segment?: string;
  thought?: string | null;
  subject_id?: string | null;
  source_agent_id?: string | null;
  organization_id?: string | null;
  scope?: string;
  status?: string;
  metadata?: MetadataMap | null;
  inserted_at?: string | null;
}
export type MemoryRecord = WireMemoryRecord;

/** Tool schema as returned by getToolSchemas(). Format depends on the requested format ("openai" or "anthropic"). */
export interface ToolSchemaRecord {
  [key: string]: unknown;
}
```

History and message protocol mapping:

```typescript
// /Users/pp/thenvoi/thenvoi-sdk-typescript/packages/sdk/src/contracts/protocols.ts:21-41
export interface HistoryConverter<T> {
  convert(raw: MetadataMap[]): T;
}

export interface PlatformMessageLike {
  id: string;
  roomId: string;
  content: string;
  senderId: string;
  senderType: string;
  senderName: string | null;
  messageType: string;
  metadata: MetadataMap;
  createdAt: Date;
}

export interface HistoryLike {
  readonly raw: MetadataMap[];
  convert<T>(converter: HistoryConverter<T>): T;
  readonly length: number;
}
```

Tool protocol mapping:

```typescript
// /Users/pp/thenvoi/thenvoi-sdk-typescript/packages/sdk/src/contracts/protocols.ts:43-95
export interface MessagingTools {
  sendMessage(content: string, mentions?: MentionInput): Promise<ToolOperationResult>;
  sendEvent(content: string, messageType: string, metadata?: MetadataMap): Promise<ToolOperationResult>;
}

export interface RoomParticipantTools {
  addParticipant(name: string, role?: string): Promise<ToolOperationResult>;
  removeParticipant(name: string): Promise<ToolOperationResult>;
  getParticipants(): Promise<ParticipantRecord[]>;
  createChatroom(taskId?: string): Promise<string>;
}

export interface PeerLookupTools {
  lookupPeers(page?: number, pageSize?: number): Promise<PaginatedList<PeerRecord>>;
}

export interface ToolSchemaProvider {
  getToolSchemas(format: "openai" | "anthropic", options?: { includeMemory?: boolean }): ToolSchemaRecord[];
  getAnthropicToolSchemas(options?: { includeMemory?: boolean }): ToolSchemaRecord[];
  getOpenAIToolSchemas(options?: { includeMemory?: boolean }): ToolSchemaRecord[];
}

export interface ToolExecutor {
  executeToolCall(toolName: string, toolArgs: MetadataMap): Promise<unknown>;
}
```

Tool executor error mapping:

```typescript
// /Users/pp/thenvoi/thenvoi-sdk-typescript/packages/sdk/src/contracts/protocols.ts:97-161
export const TOOL_EXECUTOR_ERROR_TYPES = [
  "ToolArgumentsValidationError",
  "ToolNotFoundError",
  "ToolExecutionError",
] as const;

export interface ToolExecutorError {
  ok: false;
  errorType: ToolExecutorErrorType;
  toolName: string;
  message: string;
  legacyMessage: string;
  details?: MetadataMap;
}

export function createToolExecutorError(input: {
  errorType: ToolExecutorErrorType;
  toolName: string;
  message: string;
  legacyMessage?: string;
  details?: MetadataMap;
}): ToolExecutorError {
  return {
    ok: false,
    errorType: input.errorType,
    toolName: input.toolName,
    message: input.message,
    legacyMessage: input.legacyMessage ?? input.message,
    ...(input.details ? { details: input.details } : {}),
  };
}
```

Adapter tools and framework adapter mapping:

```typescript
// /Users/pp/thenvoi/thenvoi-sdk-typescript/packages/sdk/src/contracts/protocols.ts:165-224
/** Full tool surface available to framework adapters during message handling. */
export interface AdapterToolsProtocol
  extends MessagingTools, RoomParticipantTools, ToolSchemaProvider, ToolExecutor, Partial<PeerLookupTools>, Partial<ContactTools>, Partial<MemoryTools> {
  /** Check capability flags to determine which optional tools are available. */
  readonly capabilities: Readonly<AgentToolsCapabilities>;
}

export interface AgentToolsCapabilities {
  peers: boolean;
  contacts: boolean;
  memory: boolean;
}

export const DEFAULT_AGENT_TOOLS_CAPABILITIES: AgentToolsCapabilities = {
  peers: true,
  contacts: true,
  memory: true,
};

export interface FrameworkAdapterInput {
  message: PlatformMessageLike;
  tools: AdapterToolsProtocol;
  history: HistoryLike;
  participantsMessage: string | null;
  contactsMessage: string | null;
  isSessionBootstrap: boolean;
  roomId: string;
}

/** Contract that every adapter must satisfy. Implement via {@link SimpleAdapter} for convenience. */
export interface FrameworkAdapter {
  onEvent(input: FrameworkAdapterInput): Promise<void>;
  onCleanup(roomId: string): Promise<void>;
  onStarted(agentName: string, agentDescription: string): Promise<void>;
  onRuntimeStop?(): Promise<void>;
}
```

Preprocessor mapping:

```typescript
// /Users/pp/thenvoi/thenvoi-sdk-typescript/packages/sdk/src/contracts/protocols.ts:203-239
export interface PreprocessorContext {
  roomId: string;
  hasMessage(messageId: string): boolean;
  recordMessage(message: PlatformMessageLike): void;
  getTools(): AdapterToolsProtocol;
  getRawHistory(): MetadataMap[];
  getHydratedHistory(excludeMessageId?: string): Promise<MetadataMap[]>;
  consumeParticipantsMessage(): string | null;
  consumeContactsMessage(): string | null;
  readonly isLlmInitialized: boolean;
  markLlmInitialized(): void;
  injectSystemMessage(message: string): void;
  consumeSystemMessages(): string[];
}

export interface Preprocessor<TEvent extends EventEnvelope = EventEnvelope> {
  process(context: PreprocessorContext, event: TEvent, agentId: string): Promise<FrameworkAdapterInput | null>;
}
```

History provider mapping:

```typescript
// /Users/pp/thenvoi/thenvoi-sdk-typescript/packages/sdk/src/runtime/types.ts:49-63
export class HistoryProvider {
  public readonly raw: Array<Record<string, unknown>>;

  public constructor(raw: Array<Record<string, unknown>>) {
    this.raw = raw;
  }

  public convert<T>(converter: HistoryConverter<T>): T {
    return converter.convert(this.raw);
  }

  public get length(): number {
    return this.raw.length;
  }
}
```

Error mapping:

```typescript
// /Users/pp/thenvoi/thenvoi-sdk-typescript/packages/sdk/src/core/errors.ts:1-34
export class ThenvoiSdkError extends Error {
  public constructor(message: string, cause?: unknown) {
    super(message, cause !== undefined ? { cause } : undefined);
    this.name = "ThenvoiSdkError";
  }
}

export class UnsupportedFeatureError extends ThenvoiSdkError {
  public constructor(message: string) {
    super(message);
    this.name = "UnsupportedFeatureError";
  }
}

export class ValidationError extends ThenvoiSdkError {
  public constructor(message: string, cause?: unknown) {
    super(message, cause);
    this.name = "ValidationError";
  }
}

export class TransportError extends ThenvoiSdkError {
  public constructor(message: string, cause?: unknown) {
    super(message, cause);
    this.name = "TransportError";
  }
}

export class RuntimeStateError extends ThenvoiSdkError {
  public constructor(message: string) {
    super(message);
    this.name = "RuntimeStateError";
  }
}
```

Logger mapping:

```typescript
// /Users/pp/thenvoi/thenvoi-sdk-typescript/packages/sdk/src/core/logger.ts:1-18
export interface Logger {
  debug(message: string, context?: Record<string, unknown>): void;
  info(message: string, context?: Record<string, unknown>): void;
  warn(message: string, context?: Record<string, unknown>): void;
  error(message: string, context?: Record<string, unknown>): void;
}

const noop = (): void => undefined;

export class NoopLogger implements Logger {
  public debug = noop;
  public info = noop;
  public warn = noop;
  public error = noop;
}
```

Generic adapter mapping:

```typescript
// /Users/pp/thenvoi/thenvoi-sdk-typescript/packages/sdk/src/adapters/GenericAdapter.ts:5-44
export type GenericAdapterHandler = (args: {
  message: PlatformMessage;
  tools: AdapterToolsProtocol;
  history: HistoryProvider;
  participantsMessage: string | null;
  contactsMessage: string | null;
  isSessionBootstrap: boolean;
  roomId: string;
  agentName: string;
  agentDescription: string;
}) => Promise<void>;

export class GenericAdapter extends SimpleAdapter<HistoryProvider> {
  private readonly handler: GenericAdapterHandler;

  public constructor(handler: GenericAdapterHandler) {
    super();
    this.handler = handler;
  }

  public async onMessage(...): Promise<void> {
    await this.handler({
      message,
      tools,
      history,
      participantsMessage,
      contactsMessage,
      isSessionBootstrap: context.isSessionBootstrap,
      roomId: context.roomId,
      agentName: this.agentName,
      agentDescription: this.agentDescription,
    });
  }
}
```

Public mirror names for validation: TypeScript `MetadataMap` -> Go `core.Metadata`, `ToolOperationResult` -> `core.ToolOperationResult`, `MentionReference` -> `core.MentionReference`, `PaginationMetadataLike` -> `core.PaginationMetadataLike`, `PaginatedList<TItem>` -> `core.PaginatedList[T]`, `ParticipantRecord` -> `core.ParticipantRecord`, `PeerRecord` -> `core.PeerRecord`, `ContactRecord` -> `core.ContactRecord`, `MemoryRecord` -> `core.MemoryRecord`, `ToolSchemaRecord` -> `core.ToolSchemaRecord`, `PlatformMessageLike` -> `core.PlatformMessageLike`, `HistoryProvider` -> `core.HistoryProvider`, `AdapterToolsProtocol` -> `core.AdapterToolsProtocol`, `FrameworkAdapter` -> `core.FrameworkAdapter`, `ThenvoiSdkError` -> `core.ThenvoiSdkError`, and TypeScript `GenericAdapter` -> Go `adapters.GenericAdapter`.

## Goals / Non-Goals / Constraints / Risks
Goals: define the public contracts listed in the work order, use idiomatic Go names while preserving JSON wire names where the TypeScript DTO is a wire DTO, make all async TypeScript protocol methods accept `context.Context` and return `(value, error)`, and keep tests dependency-free.

Non-goals: no REST client methods, websocket transport, platform tool execution, LLM provider adapters, console logger, or runtime session implementation.

Constraints: `AdapterToolsProtocol` in Go cannot express TypeScript `Partial<PeerLookupTools>`, `Partial<ContactTools>`, and `Partial<MemoryTools>` directly inside one interface without forcing implementations to provide optional methods. The plan is to make `AdapterToolsProtocol` require the mandatory tools plus `Capabilities() AgentToolsCapabilities`, while still exporting `PeerLookupTools`, `ContactTools`, and `MemoryTools` as separate optional interfaces that callers can type-assert when capability flags are true.

Risks: public naming must remain stable for downstream work orders, `MentionInput` is a TypeScript union and needs a Go representation that does not block either string mentions or structured mention references, and `HistoryProvider.Raw()` should avoid hidden copies because the contract says `Convert` passes the raw slice through unchanged.

## Files / Surfaces Expected To Change
- `core/dtos.go` — define `Metadata`, operation results, mention references/input, pagination, participant/peer/contact/memory/tool schema records, contact/memory request DTOs, and `PlatformMessageLike`.
- `core/protocols.go` — define tool interfaces, `AdapterToolsProtocol`, `FrameworkAdapterInput`, `PreprocessorContext`, `Preprocessor`, `FrameworkAdapter`, `AgentToolsCapabilities`, and `DefaultAgentToolsCapabilities`.
- `core/errors.go` — define typed SDK errors with `Error()`, `Name()`, and `Unwrap()` behavior plus constructors or struct fields that allow causes.
- `core/logger.go` — define `Logger` and `NoopLogger` with `Debug`, `Info`, `Warn`, and `Error` methods that do nothing.
- `core/history.go` — define `HistoryConverter[T]`, `HistoryProvider`, `NewHistoryProvider`, `Raw`, `Len`, and `Convert`.
- `adapters/generic.go` — define `GenericAdapterHandler`, handler input struct, `GenericAdapter`, lifecycle state, and no-op cleanup/runtime stop methods.
- `core/core_test.go` — cover typed errors, history conversion, noop logger, and tool executor error helpers.
- `adapters/generic_test.go` — cover `GenericAdapter` lifecycle and handler forwarding.
- `go.mod` — no planned dependency changes; modify only if Go package metadata requires it.

## Implementation Approach
Create `core` as a small, dependency-free package. Use `type Metadata map[string]any` for TypeScript `MetadataMap`, `map[string]any` fields for open-ended DTO data, pointers for nullable or optional scalars, and generic `PaginatedList[T]` with `Data []T` plus optional `Metadata *PaginationMetadataLike`.

Define `MentionInput` as a struct with `Handles []string` and `References []MentionReference` rather than `[]any`, so callers can represent both TypeScript union arms without runtime type ambiguity. Add a doc comment that states runtime serialization must choose exactly one arm and reject or define precedence when both slices are populated. Preserve snake_case JSON tags on wire-derived contact and memory fields.

Define protocol interfaces with `context.Context` as the first argument for methods that correspond to TypeScript promises. `AdapterToolsProtocol` should embed the mandatory tool groups and expose `Capabilities() AgentToolsCapabilities`; optional groups remain separately exported interfaces. `DefaultAgentToolsCapabilities` should be a function returning the default value to prevent package-level mutable state, with a doc comment that each call returns an independent struct value.

Define `HistoryProvider` with an internal raw `[]Metadata`, `Raw() []Metadata`, `Len() int`, and `Convert(converter HistoryConverter[T]) T`. `Convert` must call `converter.Convert(h.raw)` and must not clone, sort, append, or mutate the slice.

Implement typed errors as Go error structs with stable `Name() string`, `Error() string`, and `Unwrap() error`. Provide constructors for the base and derived error types where useful. `errors.Is` should see wrapped causes through `Unwrap`, and `errors.As` should match concrete SDK error pointer types.

Implement `ToolExecutorError` as a struct with `OK bool`, `ErrorType`, `ToolName`, `Message`, `LegacyMessage`, and optional `Details`. `NewToolExecutorError` should default `LegacyMessage` to `Message`, `IsToolExecutorError` should validate accepted error type strings and required fields, and `LegacyToolExecutorErrorMessage` should return the string itself for legacy string values, the structured legacy message for valid structured values, and `("", false)` or equivalent for unsupported values.

Implement `adapters.GenericAdapter` as a concrete `core.FrameworkAdapter`: store the handler, agent name, and agent description; `OnStarted` records name and description; `OnEvent` forwards message, tools, history, participant/contact messages, bootstrap flag, room ID, agent name, and description to the handler; `OnCleanup` and `OnRuntimeStop` return nil.

## Verification Strategy
Validation should run the Go build, vet, and test suite, then run the architect-required grep checks to prove the expected public interfaces and error type exist. Unit tests should cover the work-order named cases: each typed error name and cause behavior, `HistoryProvider` raw pass-through and length, `GenericAdapter` lifecycle and handler arguments, and structured plus legacy tool executor error helper behavior.

## verificationCommands
- `go test ./...`
- `grep -R "type FrameworkAdapter interface" -n core adapters`
- `grep -R "type AdapterToolsProtocol interface" -n core`
- `grep -R "type ThenvoiSdkError" -n core`
- `go build ./...`
- `go vet ./...`

## Rollback / Safety Notes
The change is additive: rollback can remove the new `core` and `adapters` files plus their tests. No network behavior, persistent storage, stdout/stderr logging, or existing client package behavior should change.

## Open Questions
None blocking. The plan assumes Go protocol methods should accept `context.Context` because the TypeScript source uses promises and existing Go client methods already use contexts.
