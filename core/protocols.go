package core

import "context"

type MessagingTools interface {
	SendMessage(ctx context.Context, content string, mentions *MentionInput) (*ToolOperationResult, error)
	SendEvent(ctx context.Context, content string, messageType string, metadata Metadata) (*ToolOperationResult, error)
}

type RoomParticipantTools interface {
	AddParticipant(ctx context.Context, name string, role *string) (*ToolOperationResult, error)
	RemoveParticipant(ctx context.Context, name string) (*ToolOperationResult, error)
	GetParticipants(ctx context.Context) ([]ParticipantRecord, error)
	CreateChatroom(ctx context.Context, taskID *string) (string, error)
}

type PeerLookupTools interface {
	LookupPeers(ctx context.Context, page *int, pageSize *int) (*PaginatedList[PeerRecord], error)
}

type ContactTools interface {
	ListContacts(ctx context.Context, args ListContactsArgs) (*PaginatedList[ContactRecord], error)
	AddContact(ctx context.Context, args AddContactArgs) (*ToolOperationResult, error)
	RemoveContact(ctx context.Context, args RemoveContactArgs) (*ToolOperationResult, error)
	ListContactRequests(ctx context.Context, args ListContactRequestsArgs) (*ContactRequestsResult, error)
	RespondContactRequest(ctx context.Context, args RespondContactRequestArgs) (*ToolOperationResult, error)
}

type MemoryTools interface {
	ListMemories(ctx context.Context, args *ListMemoriesArgs) (*PaginatedList[MemoryRecord], error)
	StoreMemory(ctx context.Context, args StoreMemoryArgs) (*MemoryRecord, error)
	GetMemory(ctx context.Context, memoryID string) (*MemoryRecord, error)
	SupersedeMemory(ctx context.Context, memoryID string) (*ToolOperationResult, error)
	ArchiveMemory(ctx context.Context, memoryID string) (*ToolOperationResult, error)
}

type ToolSchemaOptions struct {
	IncludeMemory bool
}

type ToolSchemaProvider interface {
	GetToolSchemas(format string, options *ToolSchemaOptions) []ToolSchemaRecord
	GetAnthropicToolSchemas(options *ToolSchemaOptions) []ToolSchemaRecord
	GetOpenAIToolSchemas(options *ToolSchemaOptions) []ToolSchemaRecord
}

type ToolExecutor interface {
	ExecuteToolCall(ctx context.Context, toolName string, toolArgs Metadata) (any, error)
}

type AgentToolsCapabilities struct {
	Peers    bool
	Contacts bool
	Memory   bool
}

func DefaultAgentToolsCapabilities() AgentToolsCapabilities {
	return AgentToolsCapabilities{Peers: true, Contacts: true, Memory: true}
}

// AdapterToolsProtocol contains the mandatory adapter tool groups. Optional
// groups are exported as separate interfaces and guarded by Capabilities.
type AdapterToolsProtocol interface {
	MessagingTools
	RoomParticipantTools
	ToolSchemaProvider
	ToolExecutor
	Capabilities() AgentToolsCapabilities
}

type FrameworkAdapterInput struct {
	Message             PlatformMessageLike
	Tools               AdapterToolsProtocol
	History             *HistoryProvider
	ParticipantsMessage *string
	ContactsMessage     *string
	IsSessionBootstrap  bool
	RoomID              string
}

type EventEnvelope struct {
	Type    string   `json:"type"`
	RoomID  *string  `json:"roomId"`
	Payload Metadata `json:"payload"`
	Raw     Metadata `json:"raw,omitempty"`
}

type PreprocessorContext interface {
	RoomID() string
	HasMessage(messageID string) bool
	RecordMessage(message PlatformMessageLike)
	GetTools() AdapterToolsProtocol
	GetRawHistory() []Metadata
	GetHydratedHistory(ctx context.Context, excludeMessageID *string) ([]Metadata, error)
	ConsumeParticipantsMessage() *string
	ConsumeContactsMessage() *string
	IsLLMInitialized() bool
	MarkLLMInitialized()
	InjectSystemMessage(message string)
	ConsumeSystemMessages() []string
}

type Preprocessor[T any] interface {
	Process(ctx context.Context, context PreprocessorContext, event T, agentID string) (*FrameworkAdapterInput, error)
}

type FrameworkAdapter interface {
	OnEvent(ctx context.Context, input FrameworkAdapterInput) error
	OnCleanup(ctx context.Context, roomID string) error
	OnStarted(ctx context.Context, agentName string, agentDescription string) error
	OnRuntimeStop(ctx context.Context) error
}

type ToolExecutorErrorType string

const (
	ToolArgumentsValidationError ToolExecutorErrorType = "ToolArgumentsValidationError"
	ToolNotFoundError            ToolExecutorErrorType = "ToolNotFoundError"
	ToolExecutionError           ToolExecutorErrorType = "ToolExecutionError"
)

type ToolExecutorError struct {
	OK            bool                  `json:"ok"`
	ErrorType     ToolExecutorErrorType `json:"errorType"`
	ToolName      string                `json:"toolName"`
	Message       string                `json:"message"`
	LegacyMessage string                `json:"legacyMessage"`
	Details       Metadata              `json:"details,omitempty"`
}

type ToolExecutorErrorInput struct {
	ErrorType     ToolExecutorErrorType
	ToolName      string
	Message       string
	LegacyMessage string
	Details       Metadata
}

func NewToolExecutorError(input ToolExecutorErrorInput) ToolExecutorError {
	legacyMessage := input.LegacyMessage
	if legacyMessage == "" {
		legacyMessage = input.Message
	}
	return ToolExecutorError{
		OK:            false,
		ErrorType:     input.ErrorType,
		ToolName:      input.ToolName,
		Message:       input.Message,
		LegacyMessage: legacyMessage,
		Details:       input.Details,
	}
}

func IsToolExecutorError(value any) bool {
	err, ok := value.(ToolExecutorError)
	if !ok {
		errPtr, ok := value.(*ToolExecutorError)
		if !ok || errPtr == nil {
			return false
		}
		err = *errPtr
	}
	return !err.OK && validToolExecutorErrorType(err.ErrorType) && err.ToolName != "" && err.Message != "" && err.LegacyMessage != ""
}

func LegacyToolExecutorErrorMessage(value any) (string, bool) {
	switch typed := value.(type) {
	case string:
		return typed, true
	case ToolExecutorError:
		if IsToolExecutorError(typed) {
			return typed.LegacyMessage, true
		}
	case *ToolExecutorError:
		if IsToolExecutorError(typed) {
			return typed.LegacyMessage, true
		}
	}
	return "", false
}

func validToolExecutorErrorType(errorType ToolExecutorErrorType) bool {
	switch errorType {
	case ToolArgumentsValidationError, ToolNotFoundError, ToolExecutionError:
		return true
	default:
		return false
	}
}
