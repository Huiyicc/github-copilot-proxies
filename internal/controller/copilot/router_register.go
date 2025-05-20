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

// loadConfig loads the configuration from environment variables.
func loadConfig() (*Config, error) {
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
	config, err := loadConfig()
	if err != nil {
		log.Fatal(err)
	}

	// 基础路由
	setupBasicRoutes(g, config)

	// 用户相关路由
	setupUserRoutes(g)

	// Copilot相关路由
	setupCopilotRoutes(g, config)

	// API v3相关路由
	setupV3Routes(g)
}

// setupBasicRoutes 设置基础路由
func setupBasicRoutes(g *gin.RouterGroup, config *Config) {
	g.GET("/models", createModelsHandler(config))
	g.GET("/_ping", GetPing)
	g.POST("/telemetry", PostTelemetry)
	g.GET("/agents", GetAgents)
	g.GET("/copilot_internal/user", GetCopilotInternalUser)
}

// setupUserRoutes 设置用户相关路由
func setupUserRoutes(g *gin.RouterGroup) {
	authMiddleware := middleware.AccessTokenCheckAuth()

	userGroup := g.Group("")
	userGroup.Use(authMiddleware)
	{
		userGroup.GET("/user", GetLoginUser)
		userGroup.GET("/user/orgs", GetUserOrgs)
		userGroup.GET("/api/v3/user", GetLoginUser)
		userGroup.GET("/api/v3/user/orgs", GetUserOrgs)
		userGroup.GET("/teams/:teamID/memberships/:username", GetMembership)
		userGroup.POST("/chunks", HandleChunks)
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
		completionsGroup.POST("/v1/engines/:model-name/completions", createCompletionsHandler(config))
		completionsGroup.POST("/v1/engines/copilot-codex", createCompletionsHandler(config))
		completionsGroup.POST("/chat/completions", createChatHandler(config))
		completionsGroup.POST("/agents/chat", createChatHandler(config))
		completionsGroup.POST("/v1/chat/completions", createChatHandler(config))
		completionsGroup.POST("/v1/engines/copilot-centralus-h100/speculation", createChatEditCompletionsHandler(config))
		completionsGroup.POST("/embeddings", HandleEmbeddings)
	}
}

// setupV3Routes 设置API v3相关路由
func setupV3Routes(g *gin.RouterGroup) {
	g.GET("/api/v3/meta", V3meta)
	g.GET("/api/v3/", Cliv3)
	g.GET("/", Cliv3)
}

// 处理函数生成器
func createTokenHandler(config *Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		if config.ClientType == "github" && !config.CopilotProxyAll {
			GetCopilotInternalV2Token(c)
		} else {
			GetDisguiseCopilotInternalV2Token(c)
		}
	}
}

// createCompletionsHandler 生成代码补全处理函数
func createCompletionsHandler(config *Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		if config.ClientType == "github" && config.CopilotProxyAll {
			CodexCompletions(c)
		} else {
			CodeCompletions(c)
		}
	}
}

// createChatHandler 生成聊天补全处理函数
func createChatHandler(config *Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		if config.ClientType == "github" && config.CopilotProxyAll {
			ChatsCompletions(c)
		} else {
			ChatCompletions(c)
		}
	}
}

// createChatEditCompletionsHandler 生成聊天编辑补全处理函数
func createChatEditCompletionsHandler(config *Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		if config.ClientType == "github" && config.CopilotProxyAll {
			ChatEditCompletions(c)
		} else {
			CodeCompletions(c)
		}
	}
}

// createModelsHandler 生成模型处理函数
func createModelsHandler(config *Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		if config.ClientType == "github" && config.CopilotProxyAll {
			GetCopilotModels(c)
		} else {
			GetModels(c)
		}
	}
}
