package github_auth

import (
	"encoding/json"
	"fmt"
	"github.com/gofrs/uuid"
	"github.com/gomodule/redigo/redis"
	"ripper/internal/cache"
	"strings"
)

type ClientAuthInfo struct {
	ClientId   string `json:"client_id"`
	DeviceCode string `json:"device_code"`
	UserCode   string `json:"user_code"`
	CardCode   string `json:"card_code"`
}

type ClientOAuthInfo struct {
	ClientId     string `json:"client_id" form:"client_id"`
	Code         string `json:"code" form:"code"`
	ClientSecret string `json:"client_secret" form:"client_secret"`
	Scope        string `json:"scope" form:"scope"`
}

// BindClientToCode 绑定客户端到代码
// clientId 客户端ID
// exp 过期时间
// return 用户代码, 设备代码, 错误
func BindClientToCode(clientId string, exp int) (string, string, error) {
	genCode := func() string {
		newUUID, _ := uuid.NewV4()
		uuidStr := strings.Replace(newUUID.String(), "-", "", -1)
		return uuidStr[:6]
	}
	formattedUUID := genCode()
	rep := 0
	redisKey := fmt.Sprintf("copilot.proxy.%s", formattedUUID)
	repeat, _ := cache.Exist(redisKey)
	for repeat {
		if rep > 5 {
			return "", "", fmt.Errorf("gen code error")
		}
		formattedUUID = genCode()
		redisKey = fmt.Sprintf("copilot.proxy.%s", formattedUUID)
		repeat, _ = cache.Exist(redisKey)
		rep++
	}
	devId := GenDevicesCode(40)
	authInfo := ClientAuthInfo{
		ClientId:   clientId,
		DeviceCode: devId,
		UserCode:   formattedUUID,
	}
	authInfoData, _ := json.Marshal(authInfo)
	err := cache.Set(redisKey, authInfoData, exp)
	if err != nil {
		return "", "", err
	}
	redisKey = fmt.Sprintf("copilot.proxy.map.%s", devId)
	err = cache.Set(redisKey, formattedUUID, exp)
	return formattedUUID, devId, err
}

// GetClientAuthInfoByDeviceCode 通过设备代码获取客户端授权信息
func GetClientAuthInfoByDeviceCode(deviceCode string) (*ClientAuthInfo, error) {
	redisKey := fmt.Sprintf("copilot.proxy.map.%s", deviceCode)
	userCode, err := cache.Get(redisKey)
	if err != nil {
		return nil, err
	}
	redisKey = fmt.Sprintf("copilot.proxy.%s", userCode)
	authInfoData, err := redis.Bytes(cache.Get(redisKey))
	if err != nil {
		return nil, err
	}
	authInfo := &ClientAuthInfo{}
	err = json.Unmarshal(authInfoData, &authInfo)
	return authInfo, err
}

func GetOAuthCodeInfoByClientIdAndCode(clientId string, code string) (*ClientOAuthInfo, error) {
	cacheKey := "oauth2_authorize_" + clientId
	oauthCodeData, err := redis.Bytes(cache.Get(cacheKey))
	if err != nil {
		return nil, err
	}

	var oauthCode ClientOAuthInfo
	err = json.Unmarshal(oauthCodeData, &oauthCode)
	if err != nil {
		return nil, err
	}
	if oauthCode.Code != code {
		return nil, fmt.Errorf("invalid oauth code")
	}
	return &oauthCode, nil
}

func GetClientAuthInfo(code string) (ClientAuthInfo, error) {
	redisKey := fmt.Sprintf("copilot.proxy.%s", code)
	authInfoData, err := redis.Bytes(cache.Get(redisKey))
	if err != nil {
		return ClientAuthInfo{}, err
	}
	var authInfo ClientAuthInfo
	err = json.Unmarshal(authInfoData, &authInfo)
	return authInfo, err
}

// GenDevicesCode 生成设备代码
func GenDevicesCode(codeLen int) string {
	var newUUID string
	for len(newUUID) < 64 {
		ud, _ := uuid.NewV4()
		newUUID += strings.Replace(ud.String(), "-", "", -1)
	}
	return newUUID[:codeLen]
}

// UpdateClientAuthStatusByDeviceCode 更新客户端授权码通过设备代码
func UpdateClientAuthStatusByDeviceCode(deviceCode string, cardCode string) error {
	redisKey := fmt.Sprintf("copilot.proxy.map.%s", deviceCode)
	uCode, err := cache.Get(redisKey)
	if err != nil {
		return err
	}
	redisKey = fmt.Sprintf("copilot.proxy.%s", uCode)
	authInfoData, err := redis.Bytes(cache.Get(redisKey))
	if err != nil {
		return err
	}
	authInfo := &ClientAuthInfo{}
	err = json.Unmarshal(authInfoData, &authInfo)
	if err != nil {
		return err
	}
	authInfo.CardCode = cardCode
	authInfoData, _ = json.Marshal(authInfo)
	err = cache.Set(redisKey, authInfoData, -1)
	return err
}

func RemoveClientAuthInfoByDeviceCode(deviceCode string) error {
	redisKey := fmt.Sprintf("copilot.proxy.map.%s", deviceCode)
	uCode, err := cache.Get(redisKey)
	if err != nil {
		return err
	}
	redisKey = fmt.Sprintf("copilot.proxy.%s", uCode)
	err = cache.Del(redisKey)
	if err != nil {
		return err
	}
	redisKey = fmt.Sprintf("copilot.proxy.map.%s", deviceCode)
	err = cache.Del(redisKey)
	return err
}
