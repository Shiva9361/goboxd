package internal

import (
	"crypto/rand"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
)

type Server struct {
	Runner *CodeRunner
	Router *gin.Engine
}

type ComponentStatus struct {
	Ok      bool   `json:"ok"`
	Version string `json:"version,omitempty"`
	Error   string `json:"error,omitempty"`
}

type SmokeResponse struct {
	Status          string                     `json:"status"`
	NsjailOutput    ComponentStatus            `json:"nsjail"`
	LanguagesOutput map[string]ComponentStatus `json:"languages"`
}

var activeUIDs sync.Map

// getUniqueUID generates a random UID and ensures it is not currently in use
func getUniqueUID() string {
	for { // this basically makes a tread wait until it can get a unique UID
		n, err := rand.Int(rand.Reader, big.NewInt(100000))
		if err != nil {
			log.Panicln("Failed to generate random number:", err)
		}

		uid := 100000 + n.Int64()
		uidStr := fmt.Sprintf("%d", uid)

		_, loaded := activeUIDs.LoadOrStore(uidStr, true)
		if !loaded {
			return uidStr
		}

	}
}

// releaseUID frees the UID back to the pool when the run finishes
func releaseUID(uidStr string) {
	activeUIDs.Delete(uidStr)
}

func NewServer() *Server {
	runner := &CodeRunner{}
	runner.Initialize()

	Server := &Server{
		Router: gin.Default(),
		Runner: runner,
	}

	Server.route()
	return Server
}

func (Server *Server) route() {
	Server.Router.GET("/healthz", Server.getHealth)
	Server.Router.POST("/run", Server.postRun)
	Server.Router.GET("/readyz", Server.getReady)
}

func (server *Server) getHealth(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
	})
}

// postRun gets request from the client, validates it, and then runs the code in a nsjail
func (server *Server) postRun(c *gin.Context) {
	var req ExecutionRequest

	if err := c.ShouldBindJSON(&req); err != nil {

		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request format",
			"details": err.Error(),
		})
		return
	}

	if _, ok := server.Runner.ConfigMap[req.Language]; !ok {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request",
			"details": "Lanuguage " + req.Language + " not supported",
		})
		return
	}

	workDir, err := os.MkdirTemp("", "*_goboxd")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Could not create temp dir",
		})
		return
	}

	defer os.RemoveAll(workDir) // cleanup

	currentConfig := server.Runner.Configs[server.Runner.ConfigMap[req.Language]]

	// Sanitize source and artifact file names

	if req.SourceFilename == "" {
		req.SourceFilename = currentConfig.SourceFilename
	}

	req.SourceFilename = filepath.Base(req.SourceFilename) // simple but effective ;)

	if req.ArtifactFilename == "" {
		req.ArtifactFilename = currentConfig.Artifact
	}

	req.ArtifactFilename = filepath.Base(req.ArtifactFilename)

	sourcePath := filepath.Join(workDir, req.SourceFilename)
	if err := os.WriteFile(sourcePath, []byte(req.Source), 0644); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Could not create file for execution",
		})
		return
	}

	res := ExecutionResponse{
		Status: "ok",
	}

	buildFlags := sanitizeFlags(req.Build.Flags, server.Runner.BuildFlagsLookup[server.Runner.ConfigMap[req.Language]])

	buildArgs := prepArgs(currentConfig.BuildOptions.Args, buildFlags, req.SourceFilename, req.ArtifactFilename)

	buildLimits := sanitizeLimits(currentConfig.BuildOptions.Limits, req.Build.Limits)

	// Build
	buildResult := server.Runner.ExecuteSandboxed(currentConfig.BuildOptions.Cmd, buildArgs, buildLimits, "", workDir)

	if buildResult.Status != "" {
		res.Build = &buildResult
	}

	if buildResult.Status != "ok" {
		res.Status = "build_failed"
		for range req.Tests {
			res.Tests = append(res.Tests, ExecResult{
				Status: "not_executed",
			})
		}
		c.JSON(http.StatusOK, res)
		return
	}

	runFlags := sanitizeFlags(req.Run.Flags, server.Runner.RunFlagsLookup[server.Runner.ConfigMap[req.Language]])

	runArgs := prepArgs(currentConfig.RunOptions.Args, runFlags, req.SourceFilename, req.ArtifactFilename)

	runCmd := strings.ReplaceAll(currentConfig.RunOptions.Cmd, "{{artifact}}", req.ArtifactFilename)

	runLimits := sanitizeLimits(currentConfig.RunOptions.Limits, req.Run.Limits)

	for _, input := range req.Tests {
		out := server.Runner.ExecuteSandboxed(runCmd, runArgs, runLimits, input.Stdin, workDir) // TODO: do checks on flags, on error we need to fix things
		if out.Stdout != input.ExpectedStdout {

			trimmedOutput := strings.TrimSpace(out.Stdout)

			if trimmedOutput == input.ExpectedStdout {
				out.Status = "output_whitespace_mismatch"
			} else {
				out.Status = "wrong_output"
			}

		} else {
			out.Status = "accepted"
		}
		if out.Status != "accepted" && res.Status == "ok" {
			res.Status = out.Status // the spec said first non accepted stat
		}
		res.Tests = append(res.Tests, out)
	}

	c.JSON(http.StatusOK, res)

}

func (server *Server) getReady(c *gin.Context) {
	res := SmokeResponse{
		Status:          "ok",
		LanguagesOutput: make(map[string]ComponentStatus),
	}

	workDir, err := os.MkdirTemp("", "*_goboxd")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Could not create temp dir",
		})
		return
	}

	defer os.RemoveAll(workDir)

	status, log := nsjailTest()
	if status == false {
		res.Status = "down"
		res.NsjailOutput = ComponentStatus{
			Ok:    false,
			Error: log,
		}
	} else {
		res.NsjailOutput = ComponentStatus{
			Ok:      true,
			Version: log,
		}
	}

	for _, config := range server.Runner.Configs {
		if (config.SmokeOptions.Cmd == "") || (config.SmokeOptions.Args == nil) {
			continue // skip this lang
		}
		langResponse := ComponentStatus{
			Ok: true,
		}
		if res.Status == "down" {
			langResponse.Ok = false
			langResponse.Error = "nsjail is not working, so smoke test could not be run"
			continue
		}

		out := server.Runner.ExecuteSandboxed(config.SmokeOptions.Cmd, config.SmokeOptions.Args, config.SmokeOptions.Limits, "", workDir)
		if out.Status != "ok" && out.Status != "" { // ignoring ones with no smoke test configured (Looking at you java)
			if res.Status == "ok" {
				res.Status = "degraded"
			}

			langResponse.Ok = false
			langResponse.Error = out.Stderr
		} else {
			langResponse.Version = out.Stdout
		}
		res.LanguagesOutput[config.ID] = langResponse
	}
	if res.Status != "ok" {
		c.JSON(503, res)
	} else {
		c.JSON(http.StatusOK, res)
	}
}

func nsjailTest() (bool, string) {
	cmd := exec.Command("nsjail", "--help") // there is no --version as far as i can see :(

	out, err := cmd.CombinedOutput()

	if err != nil {
		return false, "nsjail is not installed or not working properly"
	}

	return true, strings.Split(string(out), "\n")[0]
}
