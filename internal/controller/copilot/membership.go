package copilot

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func getMembership(c *gin.Context) {
	teamID := c.Param("teamID")
	username := c.Param("username")

	c.JSON(http.StatusOK, gin.H{
		"message":           "Not Found",
		"documentation_url": "https://docs.github.com/rest/teams/members#get-team-membership-for-a-user-legacy",
		"status":            "404",
		"teamID":            teamID,
		"username":          username,
	})
}