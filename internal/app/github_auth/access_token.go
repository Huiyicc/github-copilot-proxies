package github_auth

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/gofrs/uuid"
	"os"
	"sort"
	"strings"
)

func sha256Sign(data string) string {
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

func GetAccessTokenT() string {
	t, _ := uuid.NewV4()
	return t.String()
}

func jsonMap2Token(data map[string]interface{}) string {
	if len(data) == 0 {
		return ""
	}

	keys := make([]string, 0, len(data))
	for k := range data {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var sb strings.Builder
	for i, key := range keys {
		if i > 0 {
			sb.WriteString(";")
		}
		sb.WriteString(key)
		sb.WriteString("=")
		sb.WriteString(fmt.Sprintf("%v", data[key]))
	}

	return sb.String()
}

func JsonMap2SignToken(data map[string]interface{}) string {
	token := jsonMap2Token(data)
	if token == "" {
		return ""
	}

	sign := sha256Sign(token + fmt.Sprintf(";salt=%s", os.Getenv("TOKEN_SALT")))
	return token + ";8kp=1:" + sign
}
