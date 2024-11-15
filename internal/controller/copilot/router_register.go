package copilot

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"os"
	"ripper/internal/middleware"
	"strconv"
)

type Config struct {
	ClientType      string
	CopilotProxyAll bool
}

// LoadConfig loads the configuration from environment variables.
func LoadConfig() (*Config, error) {
	proxyAll, err := strconv.ParseBool(os.Getenv("COPILOT_PROXY_ALL"))
	if err != nil {
		return nil, fmt.Errorf("invalid boolean value for COPILOT_PROXY_ALL: %v", err)
	}

	return &Config{
		ClientType:      os.Getenv("COPILOT_CLIENT_TYPE"),
		CopilotProxyAll: proxyAll,
	}, nil
}

// GinApi 注册路由
func GinApi(g *gin.RouterGroup) {
	config, err := LoadConfig()
	if err != nil {
		log.Fatal(err)
	}

	// 基础路由
	setupBasicRoutes(g)

	// 用户相关路由
	setupUserRoutes(g)

	// Copilot相关路由
	setupCopilotRoutes(g, config)

	// API v3相关路由
	setupV3Routes(g)
}

// setupBasicRoutes 设置基础路由
func setupBasicRoutes(g *gin.RouterGroup) {
	g.GET("/models", models)
	g.GET("/_ping", _ping)
	g.POST("/telemetry", postTelemetry)
	g.GET("/agents", agents)
}

// setupUserRoutes 设置用户相关路由
func setupUserRoutes(g *gin.RouterGroup) {
	authMiddleware := middleware.AccessTokenCheckAuth()

	userGroup := g.Group("")
	userGroup.Use(authMiddleware)
	{
		userGroup.GET("/user", getLoginUser)
		userGroup.GET("/user/orgs", getUserOrgs)
		userGroup.GET("/api/v3/user", getLoginUser)
		userGroup.GET("/api/v3/user/orgs", getUserOrgs)
		userGroup.GET("/teams/:teamID/memberships/:username", getMembership)
	}
}

// setupCopilotRoutes 设置Copilot相关路由
func setupCopilotRoutes(g *gin.RouterGroup, config *Config) {
	tokenMiddleware := middleware.TokenCheckAuth()

	// Copilot token endpoint
	g.GET("/copilot_internal/v2/token",
		middleware.AccessTokenCheckAuth(),
		createTokenHandler(config))

	// Completions endpoints
	completionsGroup := g.Group("")
	completionsGroup.Use(tokenMiddleware)
	{
		completionsGroup.POST("/v1/engines/copilot-codex/completions", createCompletionsHandler(config))
		completionsGroup.POST("/v1/engines/copilot-codex", createCompletionsHandler(config))
		completionsGroup.POST("/chat/completions", createChatHandler(config))
		completionsGroup.POST("/v1/chat/completions", createChatHandler(config))
		completionsGroup.POST("/v1/engines/copilot-centralus-h100/speculation", createCompletionsHandler(config))
		completionsGroup.POST("/chunks", HandleChunks)
		completionsGroup.POST("/embeddings", HandleEmbeddings)
	}
}

// setupV3Routes 设置API v3相关路由
func setupV3Routes(g *gin.RouterGroup) {
	g.GET("/api/v3/meta", v3meta)
	g.GET("/api/v3/", cliv3)
	g.GET("/", cliv3)
}

// 处理函数生成器
func createTokenHandler(config *Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		if config.ClientType == "github" && !config.CopilotProxyAll {
			getCopilotInternalV2Token(c)
		} else {
			getDisguiseCopilotInternalV2Token(c)
		}
	}
}

// createCompletionsHandler 生成代码补全处理函数
func createCompletionsHandler(config *Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		if config.ClientType == "github" && config.CopilotProxyAll {
			codexCompletions(c)
		} else {
			codeCompletions(c)
		}
	}
}

// createChatHandler 生成聊天补全处理函数
func createChatHandler(config *Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		if config.ClientType == "github" && config.CopilotProxyAll {
			chatsCompletions(c)
		} else {
			chatCompletions(c)
		}
	}
}
