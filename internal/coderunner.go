package internal

import (
	"bytes"
	"fmt"
	"log"
	"os/exec"
	"strings"

	"github.com/spf13/viper"
)

// Resource is self explantory
type Resource struct {
	WallTime   int `mapstructure:"wall_time_s" json:"wall_time_s"`
	Memory     int `mapstructure:"memory_kb" json:"memory_kb"`
	MaxProcess int `mapstructure:"max_processes" json:"max_processes"`
}

// Options represents the command, arguments, resource limits, and flags for either the build or run phase of code execution.
type Options struct {
	Cmd    string   `mapstructure:"cmd"`
	Args   []string `mapstructure:"args"`
	Limits Resource `mapstructure:"limits"`
	Flags  []string `mapstructure:"flag_allowlist"`
}

// CodeConfig represents the configuration for a single programming language, including how to build and run code in that language.
type CodeConfig struct {
	ID              string  `mapstructure:"id"`
	Name            string  `mapstructure:"name"`
	Source_filename string  `mapstructure:"source"`
	Artifact        string  `mapstructure:"artifact"`
	Build_options   Options `mapstructure:"build"`
	Run_options     Options `mapstructure:"run"`
}

// CodeRunner is the main struct that holds all the config for supported languages and a map for quick lookup
type CodeRunner struct {
	Configs   []CodeConfig
	ConfigMap map[string]int // felt this is cleaner
}

// ExecutionRequest represents the incoming JSON payload.
type ExecutionRequest struct {
	Language         string      `json:"language"`
	Source           string      `json:"source"`
	SourceFilename   string      `json:"source_filename"`
	ArtifactFilename string      `json:"artifact_filename"`
	Build            PhaseConfig `json:"build"`
	Run              PhaseConfig `json:"run"`
	Tests            []TestCase  `json:"tests"`
}

// PhaseConfig holds the limits and flags for either build or run phases.
type PhaseConfig struct {
	Limits Resource `json:"limits"`
	Flags  []string `json:"flags"`
}

// TestCase represents a single stdin/stdout test pair.
type TestCase struct {
	Stdin          string `json:"stdin"`
	ExpectedStdout string `json:"expected_stdout"`
}

// ExecutionResponse is the final payload sent back to the client.
type ExecutionResponse struct {
	Status string       `json:"status"`
	Build  *ExecResult  `json:"build,omitempty"`
	Tests  []ExecResult `json:"tests"`
}

// PhaseResult represents the outcome of the compilation step.

// TestResult represents the outcome of a single runtime test case.
type ExecResult struct {
	Status       string `json:"status"`
	Stdout       string `json:"stdout"`
	Stderr       string `json:"stderr"`
	DurationMs   int64  `json:"duration_ms"`
	MemoryPeakKb int64  `json:"memory_peak_kb"`
}

// Initialize is used to load up all the config yaml
// to support all required languages
func (c *CodeRunner) Initialize() {

	viper.SetConfigName("settings")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./config")

	if err := viper.ReadInConfig(); err != nil {
		log.Panicln("Reading config file failed with: ", err)
	}

	if err := viper.UnmarshalKey("languages", &c.Configs); err != nil {
		log.Panicln("Failed to parse config file with: ", err)
	}

	c.ConfigMap = make(map[string]int)

	for i, config := range c.Configs {
		c.ConfigMap[config.ID] = i
	}

}

func ExecuteSandboxed(cmd string, args []string, resource Resource, stdInput string, workDir string) ExecResult {
	result := ExecResult{}
	if cmd == "" {
		return result
	}

	uid := getUniqueUID()
	defer releaseUID(uid)

	nsjailArgs := []string{ // TODO: migrate these args to a config file too
		"-Mo", "--quiet",
		"--user", uid, "--group", uid,
		"-E", "PATH=/usr/local/bin:/usr/bin:/bin",
		"-E", "MALLOC_ARENA_MAX=2",
		"--chroot", "/",
		"-R", "/etc",
		"-B", fmt.Sprintf("%s:/app", workDir),
		"-T", "/tmp",
		"--time_limit", fmt.Sprintf("%d", resource.WallTime),
		"--rlimit_as", fmt.Sprintf("%d", resource.Memory/1024),
		"--cwd", "/app",
		"--",
		cmd,
	}

	nsjailArgs = append(nsjailArgs, args...)

	execution := exec.Command("nsjail", nsjailArgs...)

	var stdoutBuf, stderrBuf bytes.Buffer

	if stdInput != "" {
		execution.Stdin = strings.NewReader(stdInput)
	}

	execution.Stderr = &stderrBuf
	execution.Stdout = &stdoutBuf

	err := execution.Run()

	result.Status = "ok"
	result.Stderr = stderrBuf.String()

	log.Println("err: ", result.Stderr)
	result.Stdout = stdoutBuf.String()

	if err != nil {
		result.Status = "failed"
		log.Println("Exec failed with : ", err)
		return result
	}

	return result
}
