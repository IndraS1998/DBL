package api_server

import (
	"net/http"
	"raft/state"

	"github.com/gin-gonic/gin"
)

func LeaderOnly(node *state.Node) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method != http.MethodGet {
			node.Mu.RLock()
			defer node.Mu.RUnlock()
			isLeader := node.Status == "leader"

			if !isLeader {
				c.JSON(http.StatusPermanentRedirect, gin.H{
					"error":         "This node is not the leader. Write requests must be sent to the leader.",
					"leaderAddress": node.LeaderAddress,
				})
				c.Abort()
				return
			}
		}
		c.Next()
	}
}
