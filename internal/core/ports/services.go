package ports

import "github.com/sy-software/minerva-olive/internal/core/domain"

// ConfigService wraps the methods to handle configuration operations.
type ConfigService interface {
	// CreateSet creates a new configuration set.
	CreateSet(name string) (domain.ConfigSet, error)
	// GetSet returns the configuration set with the given name.
	GetSet(name string) (domain.ConfigSet, error)
	// GetSetJson returns the configuration set as JSON bytes
	// This includes all secrets as plain text. Ready to be uses by the client.
	// If a maxAge is specified cache older than maxAge will be discarded.
	GetSetJson(name string, maxAge int) ([]byte, error)
	// GetSetNames returns the names of all configuration sets paginated.
	GetSetNames(count int, skip int) ([]string, error)
	// RenameSet renames a configuration set.
	RenameSet(name string, newName string) (domain.ConfigSet, error)
	// DeleteSet deletes a configuration set.
	DeleteSet(name string) (domain.ConfigSet, error)
	// AddItem adds a new item to the configuration set.
	AddItem(item domain.ConfigItem, setName string) (domain.ConfigSet, error)
	// UpdateItem updates an existing item in the configuration set.
	UpdateItem(item domain.ConfigItem, setName string) (domain.ConfigSet, error)
	// RemoveItem removes an item from the configuration set.
	RemoveItem(item domain.ConfigItem, setName string) (domain.ConfigSet, error)
	// SetToJson converts a configuration set to JSON bytes.
	SetToJson(set domain.ConfigSet) ([]byte, error)
}
