package api_server

import (
	"raft/state"

	"github.com/gin-gonic/gin"
)

type APIServer struct {
	Router *gin.Engine
	Node   *state.Node
}

func NewApiServer(n *state.Node) *APIServer {
	r := gin.Default()
	s := &APIServer{Router: r, Node: n}
	SetupRoutes(r, n)
	return s
}

func (s *APIServer) Run(add string) error {
	return s.Router.Run(add)
}
