package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/thesouldev/goboxd/internal"
)

func TestMain(m *testing.M) {
	// Adjust working directory so config can be found if running from /tests
	if _, err := os.Stat("config"); os.IsNotExist(err) {
		if err := os.Chdir(".."); err != nil {
			panic(err)
		}
	}
	os.Exit(m.Run())
}

type LanguageTestCase struct {
	Name             string               `json:"name"`
	Language         string               `json:"language"`
	Source           string               `json:"source"`
	SourceFilename   string               `json:"source_filename"`
	ArtifactFilename string               `json:"artifact_filename"`
	Build            internal.PhaseConfig `json:"build"`
	Run              internal.PhaseConfig `json:"run"`
	Tests            []internal.TestCase  `json:"tests"`
}

func TestLanguages(t *testing.T) {
	server := internal.NewServer()

	err := filepath.Walk("tests/testdata", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() || filepath.Ext(path) != ".json" {
			return nil
		}

		t.Run(path, func(t *testing.T) {
			data, err := os.ReadFile(path)
			assert.NoError(t, err)

			var tc LanguageTestCase
			err = json.Unmarshal(data, &tc)
			assert.NoError(t, err)

			reqBody := internal.ExecutionRequest{
				Language:         tc.Language,
				Source:           tc.Source,
				SourceFilename:   tc.SourceFilename,
				ArtifactFilename: tc.ArtifactFilename,
				Build:            tc.Build,
				Run:              tc.Run,
				Tests:            tc.Tests,
			}

			body, _ := json.Marshal(reqBody)
			req, _ := http.NewRequest("POST", "/run", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			server.Router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)

			var resp internal.ExecutionResponse
			err = json.Unmarshal(w.Body.Bytes(), &resp)
			assert.NoError(t, err)
			assert.Equal(t, "ok", resp.Status)

			if resp.Build != nil {
				assert.Equal(t, "ok", resp.Build.Status, "Build failed: %s", resp.Build.Stderr)
			}

			assert.Equal(t, len(tc.Tests), len(resp.Tests))
			for i, testResult := range resp.Tests {
				assert.Equal(t, "accepted", testResult.Status, "Test case %d failed. Stderr: %s", i, testResult.Stderr)
				assert.Equal(t, tc.Tests[i].ExpectedStdout, testResult.Stdout, "Test case %d output mismatch", i)
			}
		})
		return nil
	})

	assert.NoError(t, err)
}
