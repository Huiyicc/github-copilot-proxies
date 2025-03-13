package copilot

import (
	"context"
	"crypto/sha256"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
)

// 常量定义
const (
	defaultDimensionSize = 1536 // 默认向量维度
	defaultModelName     = "text-embedding-3-small-512"
	markdownFilePrefix   = "File: `%s`\n```shell\n"
	markdownFileSuffix   = "```"
)

// 获取向量维度大小
func getDimensionSize() int {
	dimensionSize := defaultDimensionSize
	if dimSizeStr := os.Getenv("EMBEDDING_DIMENSION_SIZE"); dimSizeStr != "" {
		if dimSize, err := strconv.Atoi(dimSizeStr); err == nil {
			dimensionSize = dimSize
		}
	}
	return dimensionSize
}

// 计算块大小
func getChunkSize() int {
	// 根据维度大小调整块大小，这里设置为维度的1.5倍左右
	return getDimensionSize() * 3 / 2
}

// ChunkRequest 表示分块请求
type ChunkRequest struct {
	Content string `json:"content" binding:"required"`
	Path    string `json:"path" binding:"required"`
	Embed   bool   `json:"embed"`
}

// Chunk 表示内容块
type Chunk struct {
	Hash      string    `json:"hash"`
	Text      string    `json:"text"`
	Range     Range     `json:"range"`
	Embedding Embedding `json:"embedding,omitempty"`
}

// Range 表示文本范围
type Range struct {
	Start int `json:"start"`
	End   int `json:"end"`
}

// Embedding 表示向量嵌入
type Embedding struct {
	Embedding []float32 `json:"embedding"`
}

// ChunkResponse 表示分块响应
type ChunkResponse struct {
	Chunks         []Chunk `json:"chunks"`
	EmbeddingModel string  `json:"embedding_model"`
}

// ChunkService 处理文本分块和嵌入的服务
type ChunkService struct {
	embeddingClient *EmbeddingClient
	modelName       string
}

// NewChunkService 创建新的分块服务
func NewChunkService() (*ChunkService, error) {
	client, err := NewEmbeddingClient(getDimensionSize())
	if err != nil {
		return nil, fmt.Errorf("failed to create embedding client: %w", err)
	}

	return &ChunkService{
		embeddingClient: client,
		modelName:       defaultModelName,
	}, nil
}

// HandleChunks 处理分块请求的HTTP处理器
func HandleChunks(c *gin.Context) {
	var req ChunkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	service, err := NewChunkService()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to initialize service: %v", err)})
		return
	}

	chunks := service.SplitIntoChunks(req.Content, req.Path)

	if req.Embed {
		if err := service.GenerateEmbeddings(c.Request.Context(), chunks); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to generate embeddings: %v", err)})
			return
		}
	}

	resp := ChunkResponse{
		Chunks:         chunks,
		EmbeddingModel: service.modelName,
	}

	c.JSON(http.StatusOK, resp)
}

// SplitIntoChunks 将内容分割成块
func (s *ChunkService) SplitIntoChunks(content, path string) []Chunk {
	var chunks []Chunk
	lines := strings.Split(content, "\n")
	chunkSize := getChunkSize()

	// 预分配切片容量，减少内存重新分配
	estimatedChunks := len(content)/chunkSize + 1
	chunks = make([]Chunk, 0, estimatedChunks)

	var sb strings.Builder
	start := 0

	for _, line := range lines {
		lineWithNewline := line + "\n"

		// 如果当前块加上新行会超过chunkSize，并且当前块不为空
		if sb.Len()+len(lineWithNewline) > chunkSize && sb.Len() > 0 {
			// 创建新的chunk
			chunkText := sb.String()
			chunk := s.createChunk(chunkText, path, start, start+len(chunkText))
			chunks = append(chunks, chunk)

			start += len(chunkText)
			sb.Reset()
			sb.WriteString(lineWithNewline)
		} else {
			sb.WriteString(lineWithNewline)
		}
	}

	// 添加最后一个chunk
	if sb.Len() > 0 {
		chunkText := sb.String()
		chunk := s.createChunk(chunkText, path, start, start+len(chunkText))
		chunks = append(chunks, chunk)
	}

	return chunks
}

// createChunk 创建一个新的内容块
func (s *ChunkService) createChunk(text, path string, start, end int) Chunk {
	// 计算文本的SHA-256哈希
	hash := sha256.Sum256([]byte(text))

	return Chunk{
		Hash: fmt.Sprintf("%x", hash),
		Text: fmt.Sprintf(markdownFilePrefix+"%s"+markdownFileSuffix, path, text),
		Range: Range{
			Start: start,
			End:   end,
		},
		Embedding: Embedding{
			Embedding: make([]float32, 0), // 初始化为空切片
		},
	}
}

// GenerateEmbeddings 为所有块生成嵌入向量
func (s *ChunkService) GenerateEmbeddings(ctx context.Context, chunks []Chunk) error {
	if len(chunks) == 0 {
		return nil
	}

	// 对于少量块，直接串行处理
	if len(chunks) <= 5 {
		return s.generateEmbeddingsSerial(ctx, chunks)
	}

	// 对于大量块，使用并行处理
	return s.generateEmbeddingsParallel(ctx, chunks)
}

// generateEmbeddingsSerial 串行生成嵌入向量
func (s *ChunkService) generateEmbeddingsSerial(ctx context.Context, chunks []Chunk) error {
	for i := range chunks {
		text := s.extractPlainText(chunks[i].Text)

		embedding, err := s.embeddingClient.GetEmbedding(ctx, text)
		if err != nil {
			return fmt.Errorf("failed to generate embedding for chunk %d: %w", i, err)
		}

		chunks[i].Embedding.Embedding = embedding
	}

	return nil
}

// generateEmbeddingsParallel 并行生成嵌入向量
func (s *ChunkService) generateEmbeddingsParallel(ctx context.Context, chunks []Chunk) error {
	var wg sync.WaitGroup
	errChan := make(chan error, len(chunks))

	// 限制并发数量，避免过多的并发请求
	semaphore := make(chan struct{}, 10)

	for i := range chunks {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()

			// 获取信号量
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			text := s.extractPlainText(chunks[idx].Text)

			embedding, err := s.embeddingClient.GetEmbedding(ctx, text)
			if err != nil {
				errChan <- fmt.Errorf("failed to generate embedding for chunk %d: %w", idx, err)
				return
			}

			chunks[idx].Embedding.Embedding = embedding
		}(i)
	}

	// 等待所有goroutine完成
	wg.Wait()
	close(errChan)

	// 检查是否有错误
	select {
	case err := <-errChan:
		if err != nil {
			return err
		}
	default:
		// 没有错误
	}

	return nil
}

// extractPlainText 从markdown格式的文本中提取纯文本
func (s *ChunkService) extractPlainText(text string) string {
	// 移除第一行 File: 标记
	if idx := strings.Index(text, "\n"); idx != -1 {
		text = text[idx+1:]
	}

	// 移除 ```shell 和结尾的 ```
	text = strings.TrimPrefix(text, "```shell\n")
	text = strings.TrimSuffix(text, "```")

	return text
}
