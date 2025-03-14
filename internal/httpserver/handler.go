package httpserver

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"s3-presigner/internal/storage"
)

type GetPresignRequest struct {
	FileID    int64 `json:"file_id"`
	ExpiresIn int64 `json:"expires_in"`
}

func GetPresignHandler(s *storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeJSONResponse(w, http.StatusMethodNotAllowed, ResponseBody{
				"statusCode": http.StatusMethodNotAllowed,
				"error":      "Method not allowed",
			})
			return
		}

		var req GetPresignRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSONResponse(w, http.StatusBadRequest, ResponseBody{
				"statusCode": http.StatusBadRequest,
				"error":      fmt.Sprintf("Invalid request body: %v", err),
			})
			return
		}

		if req.FileID == 0 || req.ExpiresIn == 0 {
			writeJSONResponse(w, http.StatusBadRequest, ResponseBody{
				"statusCode": http.StatusBadRequest,
				"error":      "Missing required fields: file_id, expires_in",
			})
			return
		}

		fileDetails, err := s.GetFileDetails(req.FileID)
		if err != nil {
			writeJSONResponse(w, http.StatusNotFound, ResponseBody{
				"statusCode": http.StatusNotFound,
				"error":      err.Error(),
			})
			return
		}

		if fileDetails.FileID == 0 || fileDetails.Region == "" || fileDetails.Bucket == "" || fileDetails.ObjectKey == "" {
			writeJSONResponse(w, http.StatusNotFound, ResponseBody{
				"statusCode": http.StatusNotFound,
				"error":      "File details not found",
			})
			return
		}

		err = s.ObjectExists(fileDetails.Region, fileDetails.Bucket, fileDetails.ObjectKey)
		if err != nil {
			writeJSONResponse(w, http.StatusNotFound, ResponseBody{
				"statusCode": http.StatusNotFound,
				"error":      err.Error(),
			})
			return
		}

		presignedURL, err := s.GetObjectPresignedURL(storage.GetObjectPresignedURLInput{
			Region:    fileDetails.Region,
			Bucket:    fileDetails.Bucket,
			ObjectKey: fileDetails.ObjectKey,
			ExpiresIn: req.ExpiresIn,
		})
		if err != nil {
			writeJSONResponse(w, http.StatusInternalServerError, ResponseBody{
				"statusCode": http.StatusInternalServerError,
				"error":      err.Error(),
			})
			return
		}

		writeJSONResponse(w, http.StatusOK, ResponseBody{
			"statusCode": http.StatusOK,
			"url":        presignedURL,
		})
	}
}

type DeletePresignRequest struct {
	FileID    int64 `json:"file_id"`
	ExpiresIn int64 `json:"expires_in"`
}

func DeletePresignHandler(s *storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeJSONResponse(w, http.StatusMethodNotAllowed, ResponseBody{
				"statusCode": http.StatusMethodNotAllowed,
				"error":      "Method not allowed",
			})
			return
		}

		var req DeletePresignRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSONResponse(w, http.StatusBadRequest, ResponseBody{
				"statusCode": http.StatusBadRequest,
				"error":      fmt.Sprintf("Invalid request body: %v", err),
			})
			return
		}

		if req.FileID == 0 || req.ExpiresIn == 0 {
			writeJSONResponse(w, http.StatusBadRequest, ResponseBody{
				"statusCode": http.StatusBadRequest,
				"error":      "Missing required fields: file_id, expires_in",
			})
			return
		}

		fileDetails, err := s.GetFileDetails(req.FileID)
		if err != nil {
			writeJSONResponse(w, http.StatusNotFound, ResponseBody{
				"statusCode": http.StatusNotFound,
				"error":      err.Error(),
			})
			return
		}

		if fileDetails.FileID == 0 || fileDetails.Region == "" || fileDetails.Bucket == "" || fileDetails.ObjectKey == "" {
			writeJSONResponse(w, http.StatusNotFound, ResponseBody{
				"statusCode": http.StatusNotFound,
				"error":      "File details not found",
			})
			return
		}

		err = s.ObjectExists(fileDetails.Region, fileDetails.Bucket, fileDetails.ObjectKey)
		if err != nil {
			writeJSONResponse(w, http.StatusNotFound, ResponseBody{
				"statusCode": http.StatusNotFound,
				"error":      err.Error(),
			})
			return
		}

		presignedURL, err := s.DeleteObjectPresignedURL(storage.DeleteObjectPresignedURLInput{
			Region:    fileDetails.Region,
			Bucket:    fileDetails.Bucket,
			ObjectKey: fileDetails.ObjectKey,
			ExpiresIn: req.ExpiresIn,
		})
		if err != nil {
			writeJSONResponse(w, http.StatusInternalServerError, ResponseBody{
				"statusCode": http.StatusInternalServerError,
				"error":      err.Error(),
			})
			return
		}

		writeJSONResponse(w, http.StatusOK, ResponseBody{
			"statusCode": http.StatusOK,
			"url":        presignedURL,
		})
	}
}

type PutPresignRequest struct {
	OrgName        string `json:"org_name"`
	ExpiresIn      int64  `json:"expires_in"`
	FileName       string `json:"file_name"`
	MimeType       string `json:"mime_type"`
	Purpose        string `json:"purpose"`
	ParentEntityID int64  `json:"parent_entity_id"`
}

func PutPresignHandler(s *storage.Storage) http.HandlerFunc {
	log.Printf("PutPresignHandler called")
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeJSONResponse(w, http.StatusMethodNotAllowed, ResponseBody{
				"statusCode": http.StatusMethodNotAllowed,
				"error":      "Method not allowed",
			})
			return
		}

		var reqBody PutPresignRequest
		bodyBytes, err := json.Marshal(r.Body)
		if err != nil {
			log.Printf("Error reading request body for logging: %v", err)
		} else {
			log.Printf("Request body: %s", string(bodyBytes))
		}
		log.Printf("Request body: %+v", reqBody)

		var req PutPresignRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSONResponse(w, http.StatusBadRequest, ResponseBody{
				"statusCode": http.StatusBadRequest,
				"error":      fmt.Sprintf("Invalid request body: %v", err),
			})
			return
		}

		if req.OrgName == "" || req.ExpiresIn == 0 || req.FileName == "" || req.MimeType == "" || req.Purpose == "" {
			writeJSONResponse(w, http.StatusBadRequest, ResponseBody{
				"statusCode": http.StatusBadRequest,
				"error":      "Missing required fields: org_name, expires_in, file_name, mime_type, purpose",
			})
			return
		}

		uploadDetails, err := s.GetFileUploadDetails(storage.FileUploadDetailsInput{
			OrgName:        req.OrgName,
			FileName:       req.FileName,
			MimeType:       req.MimeType,
			Purpose:        req.Purpose,
			ParentEntityID: req.ParentEntityID,
		})
		if err != nil {
			writeJSONResponse(w, http.StatusBadRequest, ResponseBody{
				"statusCode": http.StatusBadRequest,
				"error":      err.Error(),
			})
			return
		}

		if uploadDetails.Region == "" || uploadDetails.Bucket == "" || uploadDetails.ObjectKey == "" {
			writeJSONResponse(w, http.StatusNotFound, ResponseBody{
				"statusCode": http.StatusNotFound,
				"error":      "Failed to get upload details",
			})
			return
		}

		presignedURL, err := s.PutObjectPresignedURL(storage.PutObjectPresignedURLInput{
			Region:    uploadDetails.Region,
			Bucket:    uploadDetails.Bucket,
			ObjectKey: uploadDetails.ObjectKey,
			ExpiresIn: req.ExpiresIn,
		})
		if err != nil {
			writeJSONResponse(w, http.StatusInternalServerError, ResponseBody{
				"statusCode": http.StatusInternalServerError,
				"error":      err.Error(),
			})
			return
		}

		writeJSONResponse(w, http.StatusOK, ResponseBody{
			"statusCode": http.StatusOK,
			"url":        presignedURL,
			"object_key": uploadDetails.ObjectKey,
		})
	}
}
