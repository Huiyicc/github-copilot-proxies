package copilot

import (
	"crypto/sha256"
	"fmt"
	"github.com/gin-gonic/gin"
	"strings"
)

const (
	dimensionSize = 512 // 向量维度
	// 根据维度大小调整块大小，这里设置为维度的1.5倍左右
	chunkSize = dimensionSize * 3 / 2 // 768字符
)

type ChunkRequest struct {
	Content string `json:"content"`
	Path    string `json:"path"`
	Embed   bool   `json:"embed"`
}

type Chunk struct {
	Hash      string    `json:"hash"`
	Text      string    `json:"text"`
	Range     Range     `json:"range"`
	Embedding Embedding `json:"embedding,omitempty"`
}

type Range struct {
	Start int `json:"start"`
	End   int `json:"end"`
}

type Embedding struct {
	Embedding []float32 `json:"embedding"`
}

type ChunkResponse struct {
	Chunks         []Chunk `json:"chunks"`
	EmbeddingModel string  `json:"embedding_model"`
}

func HandleChunks(c *gin.Context) {
	var req ChunkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	chunks := splitIntoChunks(req.Content, req.Path)

	if req.Embed {
		generateEmbeddings(chunks)
	}

	resp := ChunkResponse{
		Chunks:         chunks,
		EmbeddingModel: "text-embedding-3-small-512",
	}

	c.JSON(200, resp)
}

func splitIntoChunks(content string, path string) []Chunk {
	var chunks []Chunk

	lines := strings.Split(content, "\n")
	currentChunk := ""
	start := 0

	for _, line := range lines {
		// 如果当前块加上新行会超过chunkSize，并且当前块不为空
		if len(currentChunk)+len(line)+1 > chunkSize && len(currentChunk) > 0 {
			// 创建新的chunk
			chunk := createChunk(currentChunk, path, start, start+len(currentChunk))
			chunks = append(chunks, chunk)

			start += len(currentChunk)
			currentChunk = line + "\n"
		} else {
			currentChunk += line + "\n"
		}
	}

	// 添加最后一个chunk
	if len(currentChunk) > 0 {
		chunk := createChunk(currentChunk, path, start, start+len(currentChunk))
		chunks = append(chunks, chunk)
	}

	return chunks
}

func createChunk(text string, path string, start int, end int) Chunk {
	// 计算文本的SHA-256哈希
	hash := sha256.Sum256([]byte(text))

	return Chunk{
		Hash: fmt.Sprintf("%x", hash),
		Text: fmt.Sprintf("File: `%s`\n```shell\n%s```", path, text),
		Range: Range{
			Start: start,
			End:   end,
		},
		Embedding: Embedding{
			Embedding: make([]float32, 0), // 初始化为空切片
		},
	}
}

func generateEmbeddings(chunks []Chunk) {
	for i := range chunks {
		// 提取纯文本内容(移除markdown格式)
		text := chunks[i].Text
		// 移除第一行 File: 标记
		if idx := strings.Index(text, "\n"); idx != -1 {
			text = text[idx+1:]
		}
		// 移除 ```shell 和结尾的 ```
		text = strings.TrimPrefix(text, "```shell\n")
		text = strings.TrimSuffix(text, "```")

		embedding, err := getEmbedding(text)
		if err != nil {
			// 在生产环境中应该更好地处理错误
			fmt.Printf("Failed to generate embedding for chunk %d: %v\n", i, err)
			continue
		}
		chunks[i].Embedding.Embedding = embedding
	}
}
