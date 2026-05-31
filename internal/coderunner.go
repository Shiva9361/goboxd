package internal

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

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
	Cmd            string   `mapstructure:"cmd"`
	Args           []string `mapstructure:"args"`
	Limits         Resource `mapstructure:"limits"`
	FlagsAllowList []string `mapstructure:"flag_allowlist"`
}

// CodeConfig represents the configuration for a single programming language, including how to build and run code in that language.
type CodeConfig struct {
	ID             string  `mapstructure:"id"`
	Name           string  `mapstructure:"name"`
	SourceFilename string  `mapstructure:"source"`
	Artifact       string  `mapstructure:"artifact"`
	BuildOptions   Options `mapstructure:"build"`
	RunOptions     Options `mapstructure:"run"`
	SmokeOptions   Options `mapstructure:"smoke"`
}

// CodeRunner is the main struct that holds all the config for supported languages and a map for quick lookup
type CodeRunner struct {
	Configs          []CodeConfig
	NsjailFlags      []string
	ConfigMap        map[string]int // felt this is cleaner
	BuildFlagsLookup []FlagLookup
	RunFlagsLookup   []FlagLookup
	stdlimit         int
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
	Tests  []ExecResult `json:"tests,omitempty"`
}

// TestResult represents the outcome of a single runtime test case.
type ExecResult struct {
	Status       string `json:"status"`
	Stdout       string `json:"stdout"`
	Stderr       string `json:"stderr"`
	DurationMs   int64  `json:"duration_ms"`
	MemoryPeakKb int64  `json:"memory_peak_kb"`
}

type FlagLookup struct {
	ExactMatches map[string]struct{}
	patterns     []string
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

	c.NsjailFlags = viper.GetStringSlice("nsjail_flags")

	if viper.IsSet("std_limit") {
		c.stdlimit = viper.GetInt("std_limit")
	} else {
		c.stdlimit = 10 * 1024 // 10kb should be fine?
	}

	c.ConfigMap = make(map[string]int)

	for i, config := range c.Configs {
		c.ConfigMap[config.ID] = i
		c.RunFlagsLookup = append(c.RunFlagsLookup, buildAllowlist(config.RunOptions.FlagsAllowList))
		c.BuildFlagsLookup = append(c.BuildFlagsLookup, buildAllowlist(config.BuildOptions.FlagsAllowList))
	}

}

// ExecuteSanboxed is the wrapper function for running the code in nsjail
func (c *CodeRunner) ExecuteSandboxed(uid string, cmd string, args []string, resource Resource, stdInput string, workDir string) ExecResult {
	result := ExecResult{}
	if cmd == "" {
		return result
	}

	logReader, logWriter, err := os.Pipe()

	cgroupPath, err := os.MkdirTemp("/sys/fs/cgroup", "goboxd_*")
	if err != nil {
		log.Printf("Error: Failed to create temporary cgroup directory: %v", err)
		result.Status = "internal_error"
		return result
	}

	defer func() {
		time.Sleep(50 * time.Millisecond)

		err := os.RemoveAll(cgroupPath)
		if err != nil {
			log.Printf("Warning: Failed to remove cgroup directory %s: %v", cgroupPath, err)
		}
	}()

	// nsjail r_limit on memory is on virtual memory that is why java and js were both not able to run
	// cgroups is physical memory limit and works for all languages but requires privileged nsjail :(
	// For now to let stuff work i am gonna go with cgroups and privileged nsjail
	nsjailArgs := []string{
		"--user", uid, "--group", uid,
		"-B", fmt.Sprintf("%s:/app", workDir),
		"--cgroupv2_mount", cgroupPath,
		"--time_limit", fmt.Sprintf("%d", resource.WallTime),
		"--cgroup_mem_max", fmt.Sprintf("%d", resource.Memory*1024),
		"--rlimit_nproc", fmt.Sprintf("%d", resource.MaxProcess),
	}

	nsjailArgs = append(nsjailArgs, c.NsjailFlags...)

	nsjailArgs = append(nsjailArgs, "--")

	nsjailArgs = append(nsjailArgs, cmd)

	nsjailArgs = append(nsjailArgs, args...)

	// log.Println("Running command: ", cmd, " with args: ", args, " and nsjail args: ", nsjailArgs)

	execution := exec.Command("nsjail", nsjailArgs...)

	execution.ExtraFiles = []*os.File{logWriter} // Better than simple [ heuristic for sep :)

	nsjailLogChan := make(chan string)

	go func() {
		defer logReader.Close()
		var logs strings.Builder
		scanner := bufio.NewScanner(logReader)
		for scanner.Scan() {
			log.Printf("[NSJAIL %s] %s\n", uid, scanner.Text())
			logs.WriteString(scanner.Text() + "\n")
		}
		nsjailLogChan <- logs.String()
	}()

	stdoutBuf := CappedWriter{limit: c.stdlimit}
	stderrBuf := CappedWriter{limit: c.stdlimit}

	if stdInput != "" {
		execution.Stdin = strings.NewReader(stdInput)
	}

	execution.Stderr = &stderrBuf
	execution.Stdout = &stdoutBuf

	startTime := time.Now()
	err = execution.Start()

	logWriter.Close()

	err = execution.Wait()
	result.DurationMs = time.Since(startTime).Milliseconds()

	nsjailLogs := <-nsjailLogChan

	result.Status = "ok"
	result.Stderr = stderrBuf.String()
	result.Stdout = stdoutBuf.String()

	result.MemoryPeakKb = GetPeakMemory(cgroupPath)

	if err != nil {

		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode := exitErr.ExitCode()

			if strings.Contains(nsjailLogs, "run time >= time limit") {
				result.Status = "time_exceeded"
			} else if exitCode == 139 {
				// SEGV = oom >:(
				result.Status = "memory_exceeded"
			} else {
				result.Status = "runtime_error"
			}

		} else {
			result.Status = "internal_error"
		}

		// log.Println("Exec failed with : ", err)
		return result
	}

	return result
}
