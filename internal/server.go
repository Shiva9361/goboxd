package internal

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Server struct {
	Runner *CodeRunner
	Router *gin.Engine
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

func (app *Server) getHealth(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
	})
}

func (app *Server) postRun(c *gin.Context) {

	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
	})
}
