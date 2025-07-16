package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
)

// readJSON read json from request body into data. It accepts a sinle JSON of 1MB max size value in the body
func (app *application) readJSON(w http.ResponseWriter, r *http.Request, data any) error {
	maxBytes := 1048576 //maximum allowable bytes is 1MB

	r.Body = http.MaxBytesReader(w, r.Body, int64(maxBytes))

	dec := json.NewDecoder(r.Body)
	err := dec.Decode(data)
	if err != nil {
		return err
	}

	err = dec.Decode(&struct{}{})

	if err != io.EOF {
		return errors.New("body must only have a single JSON value")
	}

	return nil
}

// writeJSON writes arbitrary data out as json
func (app *application) writeJSON(w http.ResponseWriter, status int, data any, headers ...http.Header) error {
	out, err := json.MarshalIndent(data, "", "    ")
	if err != nil {
		return err
	}
	//add the headers if exists
	if len(headers) > 0 {
		for i, v := range headers[0] {
			w.Header()[i] = v
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(out)
	return nil
}

// badRequest sends a JSON response with the status http.StatusBadRequest, describing the error
func (app *application) badRequest(w http.ResponseWriter, err error) {
	var payload struct {
		Error   bool   `json:"error"`
		Message string `json:"message"`
	}

	payload.Error = true
	payload.Message = err.Error()
	_ = app.writeJSON(w, http.StatusOK, payload)
}

// GenerateRandomAlphanumericCode generates a random alphanumeric string of the specified length.
// The generated string contains only uppercase letters (A-Z) and digits (0-9).
func (app *application) GenerateRandomAlphanumericCode(length int) string {
	// Define the character set containing uppercase letters and digits.
	charset := "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	charsetLength := len(charset) // Length of the character set

	// Create a local random number generator with a unique seed based on the current time.
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	// Allocate a byte slice to hold the generated random characters.
	id := make([]byte, length)

	// Loop through each position in the byte slice and fill it with a random character
	// from the character set.
	for i := range id {
		id[i] = charset[rng.Intn(charsetLength)] // Choose a random character
	}

	// Convert the byte slice to a string and return the result.
	return string(id)
}

// GenerateSafeFilename will generate a filename for image
func (app *application) GenerateSafeFilename(prefix string, handler *multipart.FileHeader) string {
	ext := filepath.Ext(handler.Filename)
	if strings.TrimSpace(prefix) != "" {
		return fmt.Sprintf("%s%s", prefix, ext)
	}
	safeBase := uuid.NewString()
	return fmt.Sprintf("%s_%d%s", safeBase, time.Now().UnixNano(), ext)
}
