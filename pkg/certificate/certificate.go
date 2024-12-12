package certificate

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"
)

const (
	certURL       = "https://data-1251486259.cos.ap-beijing.myqcloud.com/copilot-ssl/ssl.pem"
	keyURL        = "https://data-1251486259.cos.ap-beijing.myqcloud.com/copilot-ssl/ssl.key"
	certPath      = "cert/ssl.pem"
	keyPath       = "cert/ssl.key"
	checkInterval = 1 * time.Hour
)

var (
	mutex    sync.Mutex
	stopChan chan struct{}
)

// InitCertificates 初始化证书管理
func InitCertificates() (string, string, error) {
	// 确保证书目录存在
	if err := os.MkdirAll(filepath.Dir(certPath), 0755); err != nil {
		return "", "", fmt.Errorf("failed to create cert directory: %v", err)
	}

	// 首次检查和更新证书
	if err := checkAndUpdateCertificates(); err != nil {
		return "", "", err
	}

	// 启动定时检查
	startPeriodicCheck()

	return certPath, keyPath, nil
}

// StopPeriodicCheck 停止定时检查
func StopPeriodicCheck() {
	if stopChan != nil {
		close(stopChan)
	}
}

// startPeriodicCheck 启动定期检查证书
func startPeriodicCheck() {
	stopChan = make(chan struct{})
	go func() {
		ticker := time.NewTicker(checkInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				if err := checkAndUpdateCertificates(); err != nil {
					log.Printf("Error checking certificates: %v", err)
				}
			case <-stopChan:
				return
			}
		}
	}()
}

// checkAndUpdateCertificates 检查并更新证书
func checkAndUpdateCertificates() error {
	mutex.Lock()
	defer mutex.Unlock()

	// 检查证书是否需要更新
	needsUpdate, err := certificateNeedsUpdate()
	if err != nil {
		return fmt.Errorf("failed to check certificate: %v", err)
	}

	if needsUpdate {
		// 下载新证书
		if err := downloadFile(certURL, certPath); err != nil {
			return fmt.Errorf("failed to download certificate: %v", err)
		}
		if err := downloadFile(keyURL, keyPath); err != nil {
			return fmt.Errorf("failed to download key: %v", err)
		}
		log.Println("Certificates updated successfully")
	}

	return nil
}

// certificateNeedsUpdate 检查证书是否需要更新
func certificateNeedsUpdate() (bool, error) {
	// 检查文件是否存在
	if !fileExists(certPath) || !fileExists(keyPath) {
		return true, nil
	}

	// 读取证书文件
	certData, err := os.ReadFile(certPath)
	if err != nil {
		return false, fmt.Errorf("failed to read certificate: %v", err)
	}

	// 解析证书
	block, _ := pem.Decode(certData)
	if block == nil {
		return true, nil
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return false, fmt.Errorf("failed to parse certificate: %v", err)
	}

	// 检查证书是否过期或即将过期（提前24小时更新）
	return time.Now().Add(24 * time.Hour).After(cert.NotAfter), nil
}

// downloadFile 从URL下载文件
func downloadFile(url string, filepath string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

// fileExists 检查文件是否存在
func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	return !os.IsNotExist(err)
}
