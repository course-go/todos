package response

import (
	"encoding/json"
	"net/http"
)

type Response struct {
	Data  map[string]any `json:"data,omitempty"`
	Error string         `json:"error,omitempty"`
}

func ErrorBytes(httpCode int) []byte {
	response := Response{
		Error: http.StatusText(httpCode),
	}

	bytes, err := json.Marshal(response)
	if err != nil {
		return nil
	}

	return bytes
}

func DataBytes(name string, data any) (bytes []byte, err error) {
	response := Response{
		Data: map[string]any{
			name: data,
		},
	}

	return json.Marshal(response) //nolint: wrapcheck
}
