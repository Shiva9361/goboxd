package internal

import (
	"os"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCodeRunnerInitialize(t *testing.T) {
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)

	for {
		if _, err := os.Stat("config"); err == nil {
			break
		}
		err := os.Chdir("..")
		if err != nil {
			t.Fatalf("Could not find config directory: %v", err)
		}
	}

	runner := &CodeRunner{}
	runner.Initialize()

	assert.NotEmpty(t, runner.Configs)
	assert.NotEmpty(t, runner.ConfigMap)
	assert.NotEmpty(t, runner.NsjailFlags)
	assert.Greater(t, runner.stdlimit, 0)

	// Check if some expected languages are present
	assert.Contains(t, runner.ConfigMap, "py3")
	assert.Contains(t, runner.ConfigMap, "cpp")
}

func TestMapStatus(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		assert.Equal(t, "ok", mapStatus(nil, ""))
	})

	t.Run("Internal Error", func(t *testing.T) {
		assert.Equal(t, "internal_error", mapStatus(os.ErrNotExist, ""))
	})

	t.Run("Runtime Error", func(t *testing.T) {
		cmd := exec.Command("false")
		err := cmd.Run()
		assert.Equal(t, "runtime_error", mapStatus(err, ""))
	})

	t.Run("Time Exceeded", func(t *testing.T) {
		cmd := exec.Command("false")
		err := cmd.Run()
		assert.Equal(t, "time_exceeded", mapStatus(err, "run time >= time limit"))
	})
}
