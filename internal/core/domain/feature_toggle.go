package domain

type ToggleFlag struct {
	Status bool        `json:"status"`
	Data   interface{} `json:"data,omitempty"`
}
