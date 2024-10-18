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
	"path/filepath"
	"strconv"
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

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
		return
	}

	log.Println("Current Environment: ", os.Getenv("ENV"))
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
	certFile := os.Getenv("CERT_FILE")
	keyFile := os.Getenv("KEY_FILE")

	// 检查端口是否被占用
	checkPortAndExit(host, httpPort)
	checkPortAndExit(host, httpsPort)

	//创建一个错误组
	g, _ := errgroup.WithContext(context.Background())

	//启动HTTP服务器
	g.Go(func() error {
		httpAddr := fmt.Sprintf("%s:%d", host, httpPort)
		log.Printf("Starting HTTP server on %s\n", httpAddr)
		return r.Run(httpAddr)
	})

	//启动HTTPS服务器
	g.Go(func() error {
		httpsAddr := fmt.Sprintf("%s:%d", host, httpsPort)
		log.Printf("Starting HTTPS server on %s\n", httpsAddr)

		server := &http.Server{
			Addr:    httpsAddr,
			Handler: r,
			TLSConfig: &tls.Config{
				MinVersion: tls.VersionTLS10,
				MaxVersion: tls.VersionTLS13,
			},
			ReadTimeout:  30 * time.Second,
			WriteTimeout: 60 * time.Second,
			IdleTimeout:  120 * time.Second,
		}

		return server.ListenAndServeTLS(certFile, keyFile)
	})

	//等待所有goroutine完成
	if err := g.Wait(); err != nil {
		log.Fatal(err)
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
