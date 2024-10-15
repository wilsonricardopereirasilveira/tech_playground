package handlers

import (
	"encoding/json"
	"net/http"
)

// PingHandler responds to ping requests
func PingHandler(w http.ResponseWriter, r *http.Request) {
	response := map[string]string{"status": "ok", "message": "pong"}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
