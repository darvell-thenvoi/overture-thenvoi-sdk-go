package core

type ThenvoiSdkError struct {
	name    string
	message string
	cause   error
}

func NewThenvoiSdkError(message string, cause error) *ThenvoiSdkError {
	return &ThenvoiSdkError{name: "ThenvoiSdkError", message: message, cause: cause}
}

func (err *ThenvoiSdkError) Error() string {
	return err.message
}

func (err *ThenvoiSdkError) Name() string {
	return err.name
}

func (err *ThenvoiSdkError) Unwrap() error {
	return err.cause
}

type UnsupportedFeatureError struct {
	*ThenvoiSdkError
}

func NewUnsupportedFeatureError(message string) *UnsupportedFeatureError {
	return &UnsupportedFeatureError{ThenvoiSdkError: &ThenvoiSdkError{name: "UnsupportedFeatureError", message: message}}
}

func (err *UnsupportedFeatureError) As(target any) bool {
	return asThenvoiSdkError(target, err.ThenvoiSdkError)
}

type ValidationError struct {
	*ThenvoiSdkError
}

func NewValidationError(message string, cause error) *ValidationError {
	return &ValidationError{ThenvoiSdkError: &ThenvoiSdkError{name: "ValidationError", message: message, cause: cause}}
}

func (err *ValidationError) As(target any) bool {
	return asThenvoiSdkError(target, err.ThenvoiSdkError)
}

type TransportError struct {
	*ThenvoiSdkError
}

func NewTransportError(message string, cause error) *TransportError {
	return &TransportError{ThenvoiSdkError: &ThenvoiSdkError{name: "TransportError", message: message, cause: cause}}
}

func (err *TransportError) As(target any) bool {
	return asThenvoiSdkError(target, err.ThenvoiSdkError)
}

type RuntimeStateError struct {
	*ThenvoiSdkError
}

func NewRuntimeStateError(message string) *RuntimeStateError {
	return &RuntimeStateError{ThenvoiSdkError: &ThenvoiSdkError{name: "RuntimeStateError", message: message}}
}

func (err *RuntimeStateError) As(target any) bool {
	return asThenvoiSdkError(target, err.ThenvoiSdkError)
}

func asThenvoiSdkError(target any, err *ThenvoiSdkError) bool {
	targetErr, ok := target.(**ThenvoiSdkError)
	if !ok {
		return false
	}
	*targetErr = err
	return true
}
