package core

type Logger interface {
	Debug(message string, context Metadata)
	Info(message string, context Metadata)
	Warn(message string, context Metadata)
	Error(message string, context Metadata)
}

type NoopLogger struct{}

func (NoopLogger) Debug(string, Metadata) {}

func (NoopLogger) Info(string, Metadata) {}

func (NoopLogger) Warn(string, Metadata) {}

func (NoopLogger) Error(string, Metadata) {}
