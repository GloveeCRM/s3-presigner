package httpserver

import (
	"encoding/json"
	"log"
	"net/http"
)

const defaultResponseType = "application/json"

type ResponseBody map[string]any

func writeJSONResponse(w http.ResponseWriter, status int, data ResponseBody) {
	w.Header().Set("Content-Type", defaultResponseType)
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("Error encoding response: %v", err)
	}
}
