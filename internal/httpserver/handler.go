package httpserver

import (
	"encoding/json"
	"fmt"
	"net/http"
	"s3-presigner/internal/storage"
	"strconv"
)

type RequestBody struct {
	FileID    int64  `json:"file_id"`
	UserID    int64  `json:"user_id"`
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
			writeJSONResponse(w, http.StatusBadRequest, ErrorResponse{Error: fmt.Sprintf("Invalid request body: %v", err)})
			return
		}

		if req.FileID == 0 || req.UserID == 0 || req.Operation == "" || req.ExpiresIn == 0 {
			writeJSONResponse(w, http.StatusBadRequest, ErrorResponse{Error: "Missing required fields: file_id, user_id, operation, expires_in"})
			return
		}

		if req.Operation != "GET" && req.Operation != "PUT" && req.Operation != "DELETE" {
			writeJSONResponse(w, http.StatusBadRequest, ErrorResponse{Error: "Unsupported operation"})
			return
		}

		fileDetails, err := s.GetFileDetails(strconv.FormatInt(req.FileID, 10), strconv.FormatInt(req.UserID, 10))
		if err != nil {
			writeJSONResponse(w, http.StatusNotFound, ErrorResponse{Error: err.Error()})
			return
		}

		if fileDetails.FileID == 0 || fileDetails.Region == "" || fileDetails.Bucket == "" || fileDetails.ObjectKey == "" {
			writeJSONResponse(w, http.StatusNotFound, ErrorResponse{Error: "File details not found"})
			return
		}

		if req.Operation == "GET" || req.Operation == "DELETE" {
			err := s.ObjectExists(fileDetails.Region, fileDetails.Bucket, fileDetails.ObjectKey)
			if err != nil {
				writeJSONResponse(w, http.StatusNotFound, ErrorResponse{Error: err.Error()})
				return
			}
		}

		var presignedURL string
		var presignErr error

		switch req.Operation {
		case "GET":
			presignedURL, presignErr = s.GetObjectPresignedURL(fileDetails.Region, fileDetails.Bucket, fileDetails.ObjectKey, req.ExpiresIn)
		case "PUT":
			presignedURL, presignErr = s.PutObjectPresignedURL(fileDetails.Region, fileDetails.Bucket, fileDetails.ObjectKey, req.ExpiresIn)
		case "DELETE":
			presignedURL, presignErr = s.DeleteObjectPresignedURL(fileDetails.Region, fileDetails.Bucket, fileDetails.ObjectKey, req.ExpiresIn)
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
