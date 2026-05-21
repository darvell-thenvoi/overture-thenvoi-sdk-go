package adapters

import (
	"context"
	"errors"

	"github.com/darvell-thenvoi/overture-thenvoi-sdk-go/core"
)

type GenericAdapterHandlerInput struct {
	Message             core.PlatformMessageLike
	Tools               core.AdapterToolsProtocol
	History             *core.HistoryProvider
	ParticipantsMessage *string
	ContactsMessage     *string
	IsSessionBootstrap  bool
	RoomID              string
	AgentName           string
	AgentDescription    string
}

type GenericAdapterHandler func(ctx context.Context, input GenericAdapterHandlerInput) error

type GenericAdapter struct {
	handler          GenericAdapterHandler
	agentName        string
	agentDescription string
}

func NewGenericAdapter(handler GenericAdapterHandler) (*GenericAdapter, error) {
	if handler == nil {
		return nil, errors.New("band: generic adapter handler is required")
	}
	return &GenericAdapter{handler: handler}, nil
}

func (adapter *GenericAdapter) OnEvent(ctx context.Context, input core.FrameworkAdapterInput) error {
	return adapter.handler(ctx, GenericAdapterHandlerInput{
		Message:             input.Message,
		Tools:               input.Tools,
		History:             input.History,
		ParticipantsMessage: input.ParticipantsMessage,
		ContactsMessage:     input.ContactsMessage,
		IsSessionBootstrap:  input.IsSessionBootstrap,
		RoomID:              input.RoomID,
		AgentName:           adapter.agentName,
		AgentDescription:    adapter.agentDescription,
	})
}

func (adapter *GenericAdapter) OnCleanup(context.Context, string) error {
	return nil
}

func (adapter *GenericAdapter) OnStarted(_ context.Context, agentName string, agentDescription string) error {
	adapter.agentName = agentName
	adapter.agentDescription = agentDescription
	return nil
}

func (adapter *GenericAdapter) OnRuntimeStop(context.Context) error {
	return nil
}
