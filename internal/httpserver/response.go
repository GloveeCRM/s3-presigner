package httpserver

import (
	"encoding/json"
	"log"
	"net/http"
)

const defaultResponseType = "application/json"

type SuccessResponse struct {
	StatusCode int    `json:"statusCode"`
	URL        string `json:"url"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func writeJSONResponse(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", defaultResponseType)
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("Error encoding response: %v", err)
	}
}
