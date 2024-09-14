package util

import (
	"crypto/md5"
	"fmt"
	"hzer/pkg/crypto"
)

//CheckPassword 用于将用户输入的密码与数据库取出的密码进行比对
func CheckPassword(PlainText, SecretKey, CipherText string) bool {
	chK, _ := crypto.AesEcrypt([]byte(PlainText), []byte(SecretKey))
	return fmt.Sprintf("%x", md5.Sum(chK)) == CipherText
}

//CreatePassword 用于将用户输入的密码进行加密
func CreatePassword(SecretKey, PlainText string) string {
	chK, _ := crypto.AesEcrypt([]byte(PlainText), []byte(SecretKey))
	return fmt.Sprintf("%x", md5.Sum(chK))
}
