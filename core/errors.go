package core

type BandSdkError struct {
	name    string
	message string
	cause   error
}

func NewBandSdkError(message string, cause error) *BandSdkError {
	return &BandSdkError{name: "BandSdkError", message: message, cause: cause}
}

func (err *BandSdkError) Error() string {
	return err.message
}

func (err *BandSdkError) Name() string {
	return err.name
}

func (err *BandSdkError) Unwrap() error {
	return err.cause
}

type UnsupportedFeatureError struct {
	*BandSdkError
}

func NewUnsupportedFeatureError(message string) *UnsupportedFeatureError {
	return &UnsupportedFeatureError{BandSdkError: &BandSdkError{name: "UnsupportedFeatureError", message: message}}
}

func (err *UnsupportedFeatureError) As(target any) bool {
	return asBandSdkError(target, err.BandSdkError)
}

type ValidationError struct {
	*BandSdkError
}

func NewValidationError(message string, cause error) *ValidationError {
	return &ValidationError{BandSdkError: &BandSdkError{name: "ValidationError", message: message, cause: cause}}
}

func (err *ValidationError) As(target any) bool {
	return asBandSdkError(target, err.BandSdkError)
}

type TransportError struct {
	*BandSdkError
}

func NewTransportError(message string, cause error) *TransportError {
	return &TransportError{BandSdkError: &BandSdkError{name: "TransportError", message: message, cause: cause}}
}

func (err *TransportError) As(target any) bool {
	return asBandSdkError(target, err.BandSdkError)
}

type RuntimeStateError struct {
	*BandSdkError
}

func NewRuntimeStateError(message string) *RuntimeStateError {
	return &RuntimeStateError{BandSdkError: &BandSdkError{name: "RuntimeStateError", message: message}}
}

func (err *RuntimeStateError) As(target any) bool {
	return asBandSdkError(target, err.BandSdkError)
}

func asBandSdkError(target any, err *BandSdkError) bool {
	targetErr, ok := target.(**BandSdkError)
	if !ok {
		return false
	}
	*targetErr = err
	return true
}

type ThenvoiSdkError = BandSdkError

func NewThenvoiSdkError(message string, cause error) *ThenvoiSdkError {
	return NewBandSdkError(message, cause)
}
