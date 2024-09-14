package copilot

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"github.com/gin-gonic/gin"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

func codeCompletions(ctx *gin.Context) {
	ctxs := ctx.Request.Context()
	debounceTime, _ := strconv.Atoi(os.Getenv("COPILOT_DEBOUNCE"))
	time.Sleep(time.Duration(debounceTime) * time.Millisecond)

	if ctxs.Err() != nil {
		abortCodex(ctx, http.StatusRequestTimeout)
		return
	}

	data, err := io.ReadAll(ctx.Request.Body)
	if err != nil {
		ctx.JSON(400, gin.H{"error": "Invalid request body"})
		return
	}
	repPredata := make(map[string]interface{})
	err = json.Unmarshal(data, &repPredata)
	if err != nil {
		ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}
	repPredata["model"] = os.Getenv("CODEX_API_MODEL_NAME")
	data, _ = json.Marshal(repPredata)
	req, err := http.NewRequest("POST", os.Getenv("CODEX_API_BASE"), bytes.NewBuffer(data))
	if err != nil {
		ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+os.Getenv("CODEX_API_KEY"))

	client := &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
	resp, err := client.Do(req)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			abortCodex(ctx, http.StatusRequestTimeout)
			return
		}

		log.Println("request completions failed:", err.Error())
		abortCodex(ctx, http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	respData, err := io.ReadAll(resp.Body)
	if err != nil {
		ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}
	ctx.Data(http.StatusOK, resp.Header.Get("Content-Type"), respData)
}

func abortCodex(c *gin.Context, status int) {
	c.Header("Content-Type", "text/event-stream")

	c.String(status, "data: [DONE]\n")
	c.Abort()
}
