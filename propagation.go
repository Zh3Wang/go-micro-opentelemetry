package oteltracing

import "github.com/asim/go-micro/v3/metadata"

type metadataSupplier struct {
	metadata *metadata.Metadata
}

func (s *metadataSupplier) Get(key string) string {
	values, valid := s.metadata.Get(key)
	if !valid {
		return ""
	}
	if len(values) == 0 {
		return ""
	}
	return values
}

func (s *metadataSupplier) Set(key string, value string) {
	s.metadata.Set(key, value)
}

func (s *metadataSupplier) Keys() []string {
	return nil
}
