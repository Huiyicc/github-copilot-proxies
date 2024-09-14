package copilot

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func getLoginUser(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{
		"login":               "github",
		"id":                  9919,
		"node_id":             "DEyOk9yZ2FuaXphdGlvbjk5MTk=",
		"avatar_url":          "https://avatars.githubusercontent.com/u/9919?v=4",
		"gravatar_id":         "",
		"url":                 "https://api.github.com/users/github",
		"html_url":            "https://github.com/github",
		"followers_url":       "https://api.github.com/users/github/followers",
		"following_url":       "https://api.github.com/users/github/following{/other_user}",
		"gists_url":           "https://api.github.com/users/github/gists{/gist_id}",
		"starred_url":         "https://api.github.com/users/github/starred{/owner}{/repo}",
		"subscriptions_url":   "https://api.github.com/users/github/subscriptions",
		"organizations_url":   "https://api.github.com/users/github/orgs",
		"repos_url":           "https://api.github.com/users/github/repos",
		"events_url":          "https://api.github.com/users/github/events{/privacy}",
		"received_events_url": "https://api.github.com/users/github/received_events",
		"type":                "User",
		"site_admin":          false,
		"name":                "GitHub",
		"company":             nil,
		"blog":                "",
		"location":            "San Francisco, CA",
		"email":               nil,
		"hireable":            nil,
		"bio":                 nil,
		"twitter_username":    nil,
		"public_repos":        498,
		"public_gists":        0,
		"followers":           42848,
		"following":           0,
		"created_at":          "2008-05-11T04:37:31Z",
		"updated_at":          "2022-11-29T19:44:55Z",
	})

}
