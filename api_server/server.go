package api_server

import (
	"github.com/gin-gonic/gin"
)

type APIServer struct {
	Router *gin.Engine
}

func NewApiServer() *APIServer {
	r := gin.Default()
	s := &APIServer{Router: r}
	SetupRoutes(r)
	return s
}

func (s *APIServer) Run(add string) error {
	return s.Router.Run(add)
}
