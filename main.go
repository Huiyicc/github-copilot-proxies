package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"hzer/internal/redis"
	"hzer/internal/router"
	"log"
	"os"
	"strconv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	log.Println("当前环境: ", os.Getenv("ENV"))
	r := gin.Default()

	//初始化router
	router.NewHTTPRouter(r)

	// 初始化redis
	redis.InitRedis()

	//开启服务器
	port, _ := strconv.Atoi(os.Getenv("PORT"))
	ipAddr := fmt.Sprintf("%s:%d", os.Getenv("HOST"), port)
	_ = r.Run(ipAddr)
}
