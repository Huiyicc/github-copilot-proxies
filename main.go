package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"ripper/pkg/certificate"
	"ripper/pkg/message"
	"strconv"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"golang.org/x/sync/errgroup"
	"ripper/internal/router"
)

// 检查端口是否被占用，如果被占用则退出程序
func checkPortAndExit(host string, port int) {
	addr := fmt.Sprintf("%s:%d", host, port)
	conn, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("端口: %d 已被占用, 运行结束!", port)
	}
	conn.Close()
}

func main() {
	// 设置日志输出
	setupLogging()

	// 在非生产环境中加载 .env 文件
	if os.Getenv("ENV") != "production" {
		if err := godotenv.Load(); err != nil {
			log.Printf("Warning: Error loading .env file: %v", err)
		}
	}

	log.Println("Current Environment: ", os.Getenv("ENV"))

	// 设置默认环境变量
	initDefaultEnv()

	r := gin.Default()
	// 添加 HSTS 中间件
	r.Use(func(c *gin.Context) {
		c.Header("Strict-Transport-Security", "max-age=0")
		c.Next()
	})

	//初始化router
	router.NewHTTPRouter(r)

	//获取配置
	httpPort, _ := strconv.Atoi(os.Getenv("PORT"))
	httpsPort, _ := strconv.Atoi(os.Getenv("HTTPS_PORT"))
	host := os.Getenv("HOST")

	// 初始化证书
	certFile, keyFile, reloadChan, err := certificate.InitCertificates()
	if err != nil {
		log.Fatalf("Failed to initialize certificates: %v", err)
	}

	// 检查端口是否被占用
	checkPortAndExit(host, httpPort)
	checkPortAndExit(host, httpsPort)

	// 创建一个带取消功能的上下文
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 创建一个错误组
	g, groupCtx := errgroup.WithContext(ctx)

	// 创建一个通道来表示服务器已经启动
	serverStarted := make(chan struct{}, 2)

	// 启动HTTP服务器
	httpServer := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", host, httpPort),
		Handler: r,
	}
	g.Go(func() error {
		log.Printf("Starting HTTP server on %s\n", httpServer.Addr)
		serverStarted <- struct{}{}
		return httpServer.ListenAndServe()
	})

	// 创建一个函数来启动HTTPS服务器
	var httpsServer *http.Server
	startHTTPSServer := func() *http.Server {
		server := &http.Server{
			Addr:    fmt.Sprintf("%s:%d", host, httpsPort),
			Handler: r,
			TLSConfig: &tls.Config{
				MinVersion: tls.VersionTLS10,
				MaxVersion: tls.VersionTLS13,
			},
			ReadTimeout:  30 * time.Second,
			WriteTimeout: 60 * time.Second,
		}

		g.Go(func() error {
			log.Printf("Starting HTTPS server on %s\n", server.Addr)
			if httpsServer == nil { // 仅在首次启动时发送信号
				serverStarted <- struct{}{}
			}
			return server.ListenAndServeTLS(certFile, keyFile)
		})

		return server
	}

	// 启动初始HTTPS服务器
	httpsServer = startHTTPSServer()

	// 等待两个服务器都启动
	<-serverStarted
	<-serverStarted

	// 显示消息或消息框
	message.ShowAppLaunchMessage()

	// 监听证书更新和关闭信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// 处理证书更新和服务器关闭
	go func() {
		for {
			select {
			case <-reloadChan:
				log.Println("Certificate update detected, reloading HTTPS server...")

				// 创建关闭超时上下文
				shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)

				// 关闭当前的HTTPS服务器
				if err := httpsServer.Shutdown(shutdownCtx); err != nil {
					log.Printf("Error shutting down HTTPS server: %v", err)
				}
				shutdownCancel()

				// 启动新的HTTPS服务器
				httpsServer = startHTTPSServer()

			case <-quit:
				log.Println("Shutdown signal received, exiting...")
				cancel()
				return

			case <-groupCtx.Done():
				log.Println("Unexpected exit, trying to shutdown gracefully...")
				cancel()
				return
			}
		}
	}()

	// 等待取消信号
	<-ctx.Done()

	// 给服务器一些时间来完成正在处理的请求
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	// 优雅地关闭服务器
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		log.Printf("HTTP server Shutdown: %v", err)
	}

	if err := httpsServer.Shutdown(shutdownCtx); err != nil {
		log.Printf("HTTPS server Shutdown: %v", err)
	}

	// 等待所有 goroutine 完成
	if err := g.Wait(); err != nil && err != http.ErrServerClosed {
		log.Printf("Error during server operations: %v", err)
	}
}

func setupLogging() {
	// 创建日志目录
	logDir := "logs"
	err := os.MkdirAll(logDir, 0755)
	if err != nil {
		log.Fatal("无法创建日志目录:", err)
	}

	// 创建日志文件，使用当前日期作为文件名
	currentTime := time.Now()
	logFileName := filepath.Join(logDir, fmt.Sprintf("%s.log", currentTime.Format("2006-01-02")))
	logFile, err := os.OpenFile(logFileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		log.Fatal("无法创建日志文件:", err)
	}

	// 设置 gin 的日志输出到文件和控制台
	gin.DefaultWriter = io.MultiWriter(logFile, os.Stdout)

	// 设置标准日志输出到文件和控制台
	log.SetOutput(io.MultiWriter(logFile, os.Stdout))
	log.SetPrefix("[Copilot Proxies] ")
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
}

// initDefaultEnv 初始化默认环境变量
func initDefaultEnv() {
	if os.Getenv("COPILOT_PROXY_ALL") == "" {
		os.Setenv("COPILOT_PROXY_ALL", "false")
	}

	if os.Getenv("COPILOT_CLIENT_TYPE") == "" {
		os.Setenv("COPILOT_CLIENT_TYPE", "default")
	}

	if os.Getenv("DISGUISE_COPILOT_TOKEN_EXPIRES_AT") == "" {
		os.Setenv("DISGUISE_COPILOT_TOKEN_EXPIRES_AT", "1800")
	}

	if os.Getenv("HTTP_CLIENT_TIMEOUT") == "" {
		os.Setenv("DISGUISE_COPILOT_TOKEN_EXPIRES_AT", "60")
	}

	if os.Getenv("COPILOT_ACCOUNT_TYPE") == "" {
		os.Setenv("COPILOT_ACCOUNT_TYPE", "individual")
	}

}
