package httpx

import (
  "encoding/json"
  "net/http"
)

type APIResponse struct {
  Success bool        `json:"success"`
  Data    interface{} `json:"data,omitempty"`
  Error   *ErrorResponse `json:"error,omitempty"`
}

type ErrorResponse struct {
  Code    string `json:"code"`
  Message string `json:"message"`
  TraceID string `json:"trace_id"`
}

// WriteJSON writes a JSON APIResponse to the ResponseWriter.
func WriteJSON(w http.ResponseWriter, resp APIResponse) {
  w.Header().Set("Content-Type", "application/json")
  json.NewEncoder(w).Encode(resp)
}
