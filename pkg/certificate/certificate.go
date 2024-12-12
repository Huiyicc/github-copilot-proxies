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
	mutex             sync.Mutex
	stopChan          chan struct{}
	httpsServerReload chan struct{} // 用于通知需要重载服务器
)

// InitCertificates 初始化证书管理
func InitCertificates() (string, string, chan struct{}, error) {
	// 确保证书目录存在
	if err := os.MkdirAll(filepath.Dir(certPath), 0755); err != nil {
		return "", "", nil, fmt.Errorf("failed to create cert directory: %v", err)
	}

	httpsServerReload = make(chan struct{}, 1)

	// 首次检查和更新证书
	if err := checkAndUpdateCertificates(); err != nil {
		return "", "", nil, err
	}

	// 启动定时检查
	startPeriodicCheck()

	return certPath, keyPath, httpsServerReload, nil
}

// StopPeriodicCheck 停止定时检查
func StopPeriodicCheck() {
	if stopChan != nil {
		close(stopChan)
	}
}

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

func checkAndUpdateCertificates() error {
	mutex.Lock()
	defer mutex.Unlock()

	needsUpdate, err := certificateNeedsUpdate()
	if err != nil {
		return fmt.Errorf("failed to check certificate: %v", err)
	}

	if needsUpdate {
		if err := downloadFile(certURL, certPath); err != nil {
			return fmt.Errorf("failed to download certificate: %v", err)
		}
		if err := downloadFile(keyURL, keyPath); err != nil {
			return fmt.Errorf("failed to download key: %v", err)
		}

		// 通知需要重载服务器
		select {
		case httpsServerReload <- struct{}{}:
			log.Println("Certificates updated, triggering server reload")
		default:
			// 如果通道已满，说明已经有一个重载信号在等待处理
			log.Println("Server reload already pending")
		}
	}

	return nil
}

func certificateNeedsUpdate() (bool, error) {
	if !fileExists(certPath) || !fileExists(keyPath) {
		return true, nil
	}

	certData, err := os.ReadFile(certPath)
	if err != nil {
		return false, fmt.Errorf("failed to read certificate: %v", err)
	}

	block, _ := pem.Decode(certData)
	if block == nil {
		return true, nil
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return false, fmt.Errorf("failed to parse certificate: %v", err)
	}

	return time.Now().Add(24 * time.Hour).After(cert.NotAfter), nil
}

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

func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	return !os.IsNotExist(err)
}
