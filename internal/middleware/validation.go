package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"subscription-aggregator/internal/validation"
)

// ValidationMiddleware validates JSON request body fields
func ValidationMiddleware(allowedFields []string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Only validate POST, PUT, PATCH requests
			if r.Method != http.MethodPost && r.Method != http.MethodPut && r.Method != http.MethodPatch {
				next.ServeHTTP(w, r)
				return
			}

			// Read the entire request body
			body, err := io.ReadAll(r.Body)
			if err != nil {
				http.Error(w, "Failed to read request body", http.StatusBadRequest)
				return
			}

			// Close the original body
			r.Body.Close()

			// Create a new ReadCloser with the body data
			r.Body = io.NopCloser(bytes.NewReader(body))

			// Parse JSON for validation
			var data map[string]interface{}
			if err := json.Unmarshal(body, &data); err != nil {
				// If it's not valid JSON, let the handler deal with it
				next.ServeHTTP(w, r)
				return
			}

			// Validate fields
			if err := validation.ValidateJSONFields(data, allowedFields); err != nil {
				if validationErrors, ok := err.(validation.ValidationErrors); ok {
					var errorMessages []string
					for _, validationErr := range validationErrors {
						errorMessages = append(errorMessages, validationErr.Message)
					}
					http.Error(w, "Validation failed: "+validationErrors.Error(), http.StatusBadRequest)
					return
				}
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			// Restore body for the next handler
			r.Body = io.NopCloser(bytes.NewReader(body))
			next.ServeHTTP(w, r)
		})
	}
}
