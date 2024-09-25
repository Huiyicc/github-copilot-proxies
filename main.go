package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"golang.org/x/sync/errgroup"
	"ripper/internal/router"
)

func main() {
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
