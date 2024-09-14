package crypto

import (
	"crypto/md5"
	"encoding/hex"
)

//GetMd5 生成32位md5字串
func GetMd5(s string) string {
	if s == "" {
		return ""
	}
	h := md5.New()
	h.Write([]byte(s))
	return hex.EncodeToString(h.Sum(nil))
}
