package store

// meta is a simple key-value store for arbitrary data
type meta map[string]interface{}

// NewMetadata creates new metadata and returns it
func NewMetadata() Metadata {
	md := make(meta)

	return &md
}

// Get reads the value for the given key and returns it
func (m meta) Get(key string) interface{} {
	return m[key]
}

// Set sets the value for the given key
func (m *meta) Set(key string, val interface{}) {
	(*m)[key] = val
}
