package internal

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestPostRunValidation(t *testing.T) {
	// Adjust working directory
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)
	for {
		if _, err := os.Stat("config"); err == nil {
			break
		}
		if err := os.Chdir(".."); err != nil {
			break
		}
	}

	gin.SetMode(gin.TestMode)
	server := NewServer()

	t.Run("Invalid Language", func(t *testing.T) {
		reqBody := ExecutionRequest{
			Language: "invalid_lang",
			Source:   "print(1)",
		}
		body, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest("POST", "/run", bytes.NewBuffer(body))
		w := httptest.NewRecorder()
		server.Router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "not supported")
	})

	t.Run("Filename Sanitization", func(t *testing.T) {
		reqBody := ExecutionRequest{
			Language:       "py3",
			Source:         "print(1)",
			SourceFilename: "../../../../../../../../../../../../../tmp/goboxd_test_evil.py",
		}
		body, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest("POST", "/run", bytes.NewBuffer(body))
		w := httptest.NewRecorder()
		server.Router.ServeHTTP(w, req)
		assert.NotEqual(t, http.StatusInternalServerError, w.Code)

		var resp ExecutionResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Contains(t, []string{"ok", "internal_error"}, resp.Status)
	})
}
