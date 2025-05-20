package copilot

import (
	_ "embed"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
)

type Pong struct {
	Now    int    `json:"now"`
	Status string `json:"status"`
	Ns1    string `json:"ns1"`
}

// GetPing 模拟ping接口
func GetPing(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, Pong{
		Now:    time.Now().Second(),
		Status: "ok",
		Ns1:    "200 OK",
	})
}

// ModelsResponse 模型列表响应结构
type ModelsResponse struct {
	Data   []interface{} `json:"data"`
	Object string        `json:"object"`
}

// GetModels 获取模型列表
func GetModels(ctx *gin.Context) {
	// 从根目录下读取models.json文件
	jsonFile, err := os.Open(filepath.Join("models.json"))
	if err != nil {
		log.Printf("无法打开models.json文件: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "无法读取模型列表数据"})
		return
	}
	defer CloseIO(jsonFile)

	// 解析JSON数据
	jsonData, err := io.ReadAll(jsonFile)
	if err != nil {
		log.Printf("读取models.json内容失败: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "无法读取模型列表数据"})
		return
	}

	var modelsResponse ModelsResponse
	if err := json.Unmarshal(jsonData, &modelsResponse); err != nil {
		log.Printf("解析models.json失败: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "无法解析模型列表数据"})
		return
	}

	// 返回模型列表数据
	ctx.JSON(http.StatusOK, modelsResponse)
}

func CloseIO(c io.Closer) {
	err := c.Close()
	if nil != err {
		log.Println(err)
	}
}
