package core

type HistoryConverter interface {
	Convert(raw []Metadata) any
}

type HistoryConverterFunc func(raw []Metadata) any

func (fn HistoryConverterFunc) Convert(raw []Metadata) any {
	return fn(raw)
}

type HistoryProvider struct {
	raw []Metadata
}

func NewHistoryProvider(raw []Metadata) *HistoryProvider {
	return &HistoryProvider{raw: raw}
}

func (provider *HistoryProvider) Raw() []Metadata {
	return provider.raw
}

func (provider *HistoryProvider) Len() int {
	return len(provider.raw)
}

func (provider *HistoryProvider) Convert(converter HistoryConverter) any {
	return converter.Convert(provider.raw)
}

func ConvertHistory[T any](provider *HistoryProvider, converter func([]Metadata) T) T {
	return converter(provider.raw)
}
