package internal

import (
	"crypto/rand"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"os"
	"path/filepath"
	"sync"

	"github.com/gin-gonic/gin"
)

type Server struct {
	Runner *CodeRunner
	Router *gin.Engine
}

var activeUIDs sync.Map

// getUniqueUID generates a random UID and ensures it is not currently in use
func getUniqueUID() string {
	for {
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

	sourcePath := filepath.Join(workDir, req.SourceFilename) // TODO: Should use default if not there
	if err := os.WriteFile(sourcePath, []byte(req.Source), 0644); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Could not create file for execution",
		})
	}

	res := ExecutionResponse{
		Status: "ok",
	}

	currentConfig := server.Runner.Configs[server.Runner.ConfigMap[req.Language]]

	buildArgs := prepArgs(currentConfig.Build_options.Args, req.Build.Flags, req.SourceFilename, req.ArtifactFilename)
	// Build
	buildResult := ExecuteSandboxed(currentConfig.Build_options.Cmd, buildArgs, currentConfig.Build_options.Limits, "", workDir)

	if buildResult.Status != "" {
		res.Build = &buildResult
	}

	runArgs := prepArgs(currentConfig.Run_options.Args, req.Run.Flags, req.SourceFilename, req.ArtifactFilename)

	for _, input := range req.Tests {
		out := ExecuteSandboxed(currentConfig.Run_options.Cmd, runArgs, req.Run.Limits, input.Stdin, workDir) // TODO: do checks on flags, on error we need to fix things
		res.Tests = append(res.Tests, out)
	}

	c.JSON(http.StatusOK, res)

}
