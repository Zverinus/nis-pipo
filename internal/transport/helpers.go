package transport

import (
	"encoding/json"
	"log"
	"net/http"

	"nis-pipo/internal/middleware"
)

func ownerIDFromContext(r *http.Request) (string, bool) {
	id, _ := r.Context().Value(middleware.UserIDKey).(string)
	return id, id != ""
}

func logError(r *http.Request, msg string, err error) {
	if err != nil {
		log.Printf("ERROR %s %s: %s: %v", r.Method, r.URL.Path, msg, err)
	}
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}
