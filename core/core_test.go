package core_test

import (
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
