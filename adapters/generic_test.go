package adapters_test

import (
	"context"
	"testing"
	"time"

	"github.com/darvell-thenvoi/overture-thenvoi-sdk-go/adapters"
	"github.com/darvell-thenvoi/overture-thenvoi-sdk-go/core"
)

func TestGenericAdapterForwardsLifecycleAndEvent(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	history := core.NewHistoryProvider([]core.Metadata{{"id": "m1"}})
	participantsMessage := "participants"
	contactsMessage := "contacts"
	tools := fakeTools{}
	message := core.PlatformMessageLike{
		ID:          "msg-1",
		RoomID:      "room-1",
		Content:     "hello",
		SenderID:    "agent-1",
		SenderType:  "agent",
		MessageType: "message",
		Metadata:    core.Metadata{"key": "value"},
		CreatedAt:   time.Unix(1, 0),
	}

	var got adapters.GenericAdapterHandlerInput
	adapter, err := adapters.NewGenericAdapter(func(_ context.Context, input adapters.GenericAdapterHandlerInput) error {
		got = input
		return nil
	})
	if err != nil {
		t.Fatalf("NewGenericAdapter returned error: %v", err)
	}

	if err := adapter.OnStarted(ctx, "agent name", "agent description"); err != nil {
		t.Fatalf("OnStarted returned error: %v", err)
	}
	if err := adapter.OnEvent(ctx, core.FrameworkAdapterInput{
		Message:             message,
		Tools:               tools,
		History:             history,
		ParticipantsMessage: &participantsMessage,
		ContactsMessage:     &contactsMessage,
		IsSessionBootstrap:  true,
		RoomID:              "room-1",
	}); err != nil {
		t.Fatalf("OnEvent returned error: %v", err)
	}

	if got.Message.ID != "msg-1" || got.Tools == nil || got.History != history {
		t.Fatalf("forwarded input = %+v", got)
	}
	if got.ParticipantsMessage == nil || *got.ParticipantsMessage != "participants" {
		t.Fatalf("participants message = %v", got.ParticipantsMessage)
	}
	if got.ContactsMessage == nil || *got.ContactsMessage != "contacts" {
		t.Fatalf("contacts message = %v", got.ContactsMessage)
	}
	if !got.IsSessionBootstrap || got.RoomID != "room-1" || got.AgentName != "agent name" || got.AgentDescription != "agent description" {
		t.Fatalf("lifecycle fields = %+v", got)
	}
	if err := adapter.OnCleanup(ctx, "room-1"); err != nil {
		t.Fatalf("OnCleanup returned error: %v", err)
	}
	if err := adapter.OnRuntimeStop(ctx); err != nil {
		t.Fatalf("OnRuntimeStop returned error: %v", err)
	}
}

func TestGenericAdapterRequiresHandler(t *testing.T) {
	t.Parallel()

	if _, err := adapters.NewGenericAdapter(nil); err == nil {
		t.Fatalf("NewGenericAdapter accepted nil handler")
	}
}

type fakeTools struct{}

func (fakeTools) SendMessage(context.Context, string, *core.MentionInput) (*core.ToolOperationResult, error) {
	return &core.ToolOperationResult{}, nil
}

func (fakeTools) SendEvent(context.Context, string, string, core.Metadata) (*core.ToolOperationResult, error) {
	return &core.ToolOperationResult{}, nil
}

func (fakeTools) AddParticipant(context.Context, string, *string) (*core.ToolOperationResult, error) {
	return &core.ToolOperationResult{}, nil
}

func (fakeTools) RemoveParticipant(context.Context, string) (*core.ToolOperationResult, error) {
	return &core.ToolOperationResult{}, nil
}

func (fakeTools) GetParticipants(context.Context) ([]core.ParticipantRecord, error) {
	return nil, nil
}

func (fakeTools) CreateChatroom(context.Context, *string) (string, error) {
	return "room-1", nil
}

func (fakeTools) GetToolSchemas(string, *core.ToolSchemaOptions) []core.ToolSchemaRecord {
	return nil
}

func (fakeTools) GetAnthropicToolSchemas(*core.ToolSchemaOptions) []core.ToolSchemaRecord {
	return nil
}

func (fakeTools) GetOpenAIToolSchemas(*core.ToolSchemaOptions) []core.ToolSchemaRecord {
	return nil
}

func (fakeTools) ExecuteToolCall(context.Context, string, core.Metadata) (any, error) {
	return nil, nil
}

func (fakeTools) Capabilities() core.AgentToolsCapabilities {
	return core.DefaultAgentToolsCapabilities()
}
