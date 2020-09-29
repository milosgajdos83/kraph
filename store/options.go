package store

import (
	"github.com/milosgajdos/kraph/store/attrs"
	"github.com/milosgajdos/kraph/store/metadata"
)

const (
	DefaultEdgeWeight = 1.0
)

// Options are store options
type Options struct {
	Attrs    Attrs
	Metadata Metadata
	Weight   float64
}

// Option sets options
type Option func(*Options)

// NewOptions returns empty options
func NewOptions() Options {
	return Options{
		Metadata: metadata.New(),
		Attrs:    attrs.New(),
		Weight:   DefaultEdgeWeight,
	}
}

// Meta sets entity metadata
func Meta(m Metadata) Option {
	return func(o *Options) {
		o.Metadata = m
	}
}

// Attributes sets entity attributes
func Attributes(a Attrs) Option {
	return func(o *Options) {
		o.Attrs = a
	}
}

// Weight returns entity weight
func Weight(w float64) Option {
	return func(o *Options) {
		o.Weight = w
	}
}
