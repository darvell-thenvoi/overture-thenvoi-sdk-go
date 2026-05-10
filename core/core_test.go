package core_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/darvell-thenvoi/overture-thenvoi-sdk-go/core"
)

func TestTypedErrorsExposeNamesAndCauses(t *testing.T) {
	t.Parallel()

	cause := errors.New("socket closed")
	err := core.NewTransportError("transport failed", cause)

	if err.Name() != "TransportError" {
		t.Fatalf("Name() = %q", err.Name())
	}
	if err.Error() != "transport failed" {
		t.Fatalf("Error() = %q", err.Error())
	}
	if !errors.Is(err, cause) {
		t.Fatalf("errors.Is did not match wrapped cause")
	}

	var sdkErr *core.ThenvoiSdkError
	if !errors.As(err, &sdkErr) {
		t.Fatalf("errors.As did not match ThenvoiSdkError")
	}

	if got := core.NewUnsupportedFeatureError("missing").Name(); got != "UnsupportedFeatureError" {
		t.Fatalf("unsupported feature name = %q", got)
	}
	if got := core.NewValidationError("bad input", cause).Name(); got != "ValidationError" {
		t.Fatalf("validation name = %q", got)
	}
	if got := core.NewRuntimeStateError("stopped").Name(); got != "RuntimeStateError" {
		t.Fatalf("runtime state name = %q", got)
	}
}

func TestHistoryProviderPassesRawSlice(t *testing.T) {
	t.Parallel()

	raw := []core.Metadata{{"id": "m1"}, {"id": "m2"}}
	history := core.NewHistoryProvider(raw)

	if history.Len() != 2 {
		t.Fatalf("Len() = %d", history.Len())
	}
	if &history.Raw()[0] != &raw[0] {
		t.Fatalf("Raw() did not return the original backing data")
	}

	got := history.Convert(core.HistoryConverterFunc(func(converted []core.Metadata) any {
		if &converted[0] != &raw[0] {
			t.Fatalf("Convert received a copied slice")
		}
		return converted[1]["id"]
	}))
	if got != "m2" {
		t.Fatalf("Convert() = %v", got)
	}

	typed := core.ConvertHistory(history, func(converted []core.Metadata) int {
		return len(converted)
	})
	if typed != 2 {
		t.Fatalf("ConvertHistory() = %d", typed)
	}
}

func TestNoopLogger(t *testing.T) {
	t.Parallel()

	var logger core.Logger = core.NoopLogger{}
	logger.Debug("debug", core.Metadata{"id": "1"})
	logger.Info("info", nil)
	logger.Warn("warn", nil)
	logger.Error("error", nil)
}

func TestToolExecutorErrorHelpers(t *testing.T) {
	t.Parallel()

	err := core.NewToolExecutorError(core.ToolExecutorErrorInput{
		ErrorType: core.ToolNotFoundError,
		ToolName:  "lookup",
		Message:   "tool not found",
		Details:   core.Metadata{"candidate": "search"},
	})

	if err.OK {
		t.Fatalf("OK = true")
	}
	if err.LegacyMessage != "tool not found" {
		t.Fatalf("LegacyMessage = %q", err.LegacyMessage)
	}
	if !core.IsToolExecutorError(err) {
		t.Fatalf("IsToolExecutorError returned false")
	}
	if msg, ok := core.LegacyToolExecutorErrorMessage(&err); !ok || msg != "tool not found" {
		t.Fatalf("LegacyToolExecutorErrorMessage() = %q, %v", msg, ok)
	}
	if msg, ok := core.LegacyToolExecutorErrorMessage("legacy failure"); !ok || msg != "legacy failure" {
		t.Fatalf("legacy string message = %q, %v", msg, ok)
	}
	if core.IsToolExecutorError(core.ToolExecutorError{ErrorType: "Other", ToolName: "lookup", Message: "bad", LegacyMessage: "bad"}) {
		t.Fatalf("invalid error type passed validation")
	}
}

func TestDefaultAgentToolsCapabilitiesReturnsEnabledValue(t *testing.T) {
	t.Parallel()

	got := core.DefaultAgentToolsCapabilities()
	if !got.Peers || !got.Contacts || !got.Memory {
		t.Fatalf("capabilities = %+v", got)
	}
}

func TestContactRequestsResultUsesRequestRecordShapes(t *testing.T) {
	t.Parallel()

	data := []byte(`{"received":[{"id":"req-1","status":"pending","message":"hello","inserted_at":"2026-01-02T03:04:05Z","from_handle":"@agent","from_name":"Agent"}],"sent":[{"id":"req-2","status":"accepted","message":"hi","inserted_at":"2026-01-03T03:04:05Z","to_handle":"@user","to_name":"User"}],"metadata":{"page":1}}`)

	var got core.ContactRequestsResult
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("Unmarshal returned error: %v", err)
	}
	if len(got.Received) != 1 || got.Received[0].Status == nil || *got.Received[0].Status != "pending" {
		t.Fatalf("received status = %+v", got.Received)
	}
	if got.Received[0].Message == nil || *got.Received[0].Message != "hello" {
		t.Fatalf("received message = %+v", got.Received)
	}
	if got.Received[0].FromHandle == nil || *got.Received[0].FromHandle != "@agent" {
		t.Fatalf("from handle = %+v", got.Received)
	}
	if len(got.Sent) != 1 || got.Sent[0].ToName == nil || *got.Sent[0].ToName != "User" {
		t.Fatalf("sent to name = %+v", got.Sent)
	}
}

func TestMemoryToolsContractIncludesUpstreamMethods(t *testing.T) {
	t.Parallel()

	var tools core.MemoryTools = fakeMemoryTools{}
	ctx := context.Background()
	filter := core.ListMemoriesArgs{}

	if _, err := tools.ListMemories(ctx, &filter); err != nil {
		t.Fatalf("ListMemories returned error: %v", err)
	}
	if _, err := tools.StoreMemory(ctx, core.StoreMemoryArgs{
		Content: "remember this",
		System:  core.MemorySystemLongTerm,
		Type:    core.MemoryTypeSemantic,
		Segment: core.MemorySegmentAgent,
		Thought: "test",
	}); err != nil {
		t.Fatalf("StoreMemory returned error: %v", err)
	}
	if _, err := tools.GetMemory(ctx, "mem-1"); err != nil {
		t.Fatalf("GetMemory returned error: %v", err)
	}
	if _, err := tools.SupersedeMemory(ctx, "mem-1"); err != nil {
		t.Fatalf("SupersedeMemory returned error: %v", err)
	}
	if _, err := tools.ArchiveMemory(ctx, "mem-1"); err != nil {
		t.Fatalf("ArchiveMemory returned error: %v", err)
	}
}

type fakeMemoryTools struct{}

func (fakeMemoryTools) ListMemories(_ context.Context, _ *core.ListMemoriesArgs) (*core.PaginatedList[core.MemoryRecord], error) {
	return &core.PaginatedList[core.MemoryRecord]{}, nil
}

func (fakeMemoryTools) StoreMemory(_ context.Context, _ core.StoreMemoryArgs) (*core.MemoryRecord, error) {
	return &core.MemoryRecord{}, nil
}

func (fakeMemoryTools) GetMemory(_ context.Context, _ string) (*core.MemoryRecord, error) {
	return &core.MemoryRecord{}, nil
}

func (fakeMemoryTools) SupersedeMemory(_ context.Context, _ string) (*core.ToolOperationResult, error) {
	return &core.ToolOperationResult{}, nil
}

func (fakeMemoryTools) ArchiveMemory(_ context.Context, _ string) (*core.ToolOperationResult, error) {
	return &core.ToolOperationResult{}, nil
}
