package domain

// ToggleFlag represents a stored flag inside the feature flag server
type ToggleFlag struct {
	// Whether this flag is on or off
	Status bool `json:"status"`
	// Associated data to be used by the consumer to act based on the flag status
	Data interface{} `json:"data,omitempty"`
}
