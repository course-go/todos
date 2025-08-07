package health

type Report struct {
	Service    string                     `json:"service"`
	Version    string                     `json:"version"`
	Health     Health                     `json:"health"`
	Components map[string]ComponentHealth `json:"components,omitempty"`
}
