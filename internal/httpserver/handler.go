package httpserver

import (
	"encoding/json"
	"net/http"
	"s3-presigner/internal/storage"
)

type RequestBody struct {
	Region    string `json:"region"`
	Bucket    string `json:"bucket"`
	ObjectKey string `json:"object_key"`
	Operation string `json:"operation"`
	ExpiresIn int64  `json:"expires_in"`
}

func PresignHandler(s *storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeJSONResponse(w, http.StatusMethodNotAllowed, ErrorResponse{Error: "Method not allowed"})
			return
		}

		var req RequestBody
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSONResponse(w, http.StatusBadRequest, ErrorResponse{Error: "Invalid request body"})
			return
		}

		if req.Region == "" || req.Bucket == "" || req.ObjectKey == "" || req.Operation == "" || req.ExpiresIn == 0 {
			writeJSONResponse(w, http.StatusBadRequest,
				ErrorResponse{Error: "Missing required fields: region, bucket, object_key, operation, expires_in"})
			return
		}

		if req.Operation != "GET" && req.Operation != "DELETE" && req.Operation != "PUT" {
			writeJSONResponse(w, http.StatusBadRequest, ErrorResponse{Error: "Unsupported operation"})
			return
		}

		if req.Operation == "GET" || req.Operation == "DELETE" {
			if err := s.ObjectExists(req.Region, req.Bucket, req.ObjectKey); err != nil {
				writeJSONResponse(w, http.StatusNotFound, ErrorResponse{Error: err.Error()})
				return
			}
		}

		var presignedURL string
		var presignErr error

		switch req.Operation {
		case "GET":
			presignedURL, presignErr = s.GetObjectPresignedURL(req.Region, req.Bucket, req.ObjectKey, req.ExpiresIn)
		case "PUT":
			presignedURL, presignErr = s.PutObjectPresignedURL(req.Region, req.Bucket, req.ObjectKey, req.ExpiresIn)
		case "DELETE":
			presignedURL, presignErr = s.DeleteObjectPresignedURL(req.Region, req.Bucket, req.ObjectKey, req.ExpiresIn)
		}

		if presignErr != nil {
			writeJSONResponse(w, http.StatusInternalServerError, ErrorResponse{Error: presignErr.Error()})
			return
		}

		writeJSONResponse(w, http.StatusOK, SuccessResponse{
			StatusCode: http.StatusOK,
			URL:        presignedURL,
		})
	}
}
