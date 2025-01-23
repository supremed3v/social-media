package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/go-playground/validator/v10"
)

var Validate *validator.Validate

func init() {
	Validate = validator.New(validator.WithRequiredStructEnabled())
}

func writeJSON(w http.ResponseWriter, status int, data any) error {
	w.Header().Set("Content-Type", "application/json")

	w.WriteHeader(status)

	return json.NewEncoder(w).Encode(data)

}

func readJSON(w http.ResponseWriter, r *http.Request, data any) error {
	maxBytes := 1_048_578 // 1mb
	r.Body = http.MaxBytesReader(w, r.Body, int64(maxBytes))
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	return decoder.Decode(data)
}

func readFile(r *http.Request, fieldName string, maxBytes int64) ([]byte, string, string, error) {
	// Parse the multipart form with the specified memory limit
	err := r.ParseMultipartForm(maxBytes)
	if err != nil {
		return nil, "", "", fmt.Errorf("failed to parse multipart form: %w", err)
	}

	// Retrieve the file from the form field
	file, header, err := r.FormFile(fieldName)
	if err != nil {
		return nil, "", "", fmt.Errorf("failed to retrieve file: %w", err)
	}
	defer file.Close()

	// Check file size
	if header.Size > maxBytes {
		return nil, "", "", fmt.Errorf("file size exceeds the limit of %d bytes", maxBytes)
	}

	// Read the file bytes
	fileBytes, err := io.ReadAll(file)
	if err != nil {
		return nil, "", "", fmt.Errorf("failed to read file: %w", err)
	}

	// Detect the file's content type
	fileType := http.DetectContentType(fileBytes)

	return fileBytes, header.Filename, fileType, nil
}

func writeJSONError(w http.ResponseWriter, status int, message string) error {
	type envelope struct {
		Error string `json:"error"`
	}

	return writeJSON(w, status, &envelope{Error: message})

}

func (app *application) jsonResponse(w http.ResponseWriter, status int, data any) error {
	type envelope struct {
		Data any `json:"data"`
	}

	return writeJSON(w, status, &envelope{Data: data})
}
