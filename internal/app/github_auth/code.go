package github_auth

import (
	"encoding/json"
	"fmt"
	"github.com/gofrs/uuid"
	"github.com/gomodule/redigo/redis"
	redisPkg "ripper/internal/redis"
	"strings"
)

type ClientAuthInfo struct {
	ClientId   string `json:"client_id"`
	DeviceCode string `json:"device_code"`
	UserCode   string `json:"user_code"`
	CardCode   string `json:"card_code"`
}

// BindClientToCode 绑定客户端到代码
// clientId 客户端ID
// exp 过期时间
// return 用户代码, 设备代码, 错误
func BindClientToCode(clientId string, exp int) (string, string, error) {
	rc := redisPkg.GetCoon()
	genCode := func() string {
		newUUID, _ := uuid.NewV4()
		uuidStr := strings.Replace(newUUID.String(), "-", "", -1)
		return uuidStr[:6]
	}
	formattedUUID := genCode()
	rep := 0
	redisKey := fmt.Sprintf("copilot.proxy.%s", formattedUUID)
	repeat, _ := redis.Bool(rc.Do("EXISTS", redisKey))
	for repeat {
		if rep > 5 {
			return "", "", fmt.Errorf("gen code error")
		}
		formattedUUID = genCode()
		redisKey = fmt.Sprintf("copilot.proxy.%s", formattedUUID)
		repeat, _ = redis.Bool(rc.Do("EXISTS", redisKey))
		rep++
	}
	devId := GenDevicesCode(40)
	authInfo := ClientAuthInfo{
		ClientId:   clientId,
		DeviceCode: devId,
		UserCode:   formattedUUID,
	}
	authInfoData, _ := json.Marshal(authInfo)
	_, err := rc.Do("set", redisKey, authInfoData, "EX", exp)
	if err != nil {
		return "", "", err
	}
	redisKey = fmt.Sprintf("copilot.proxy.map.%s", devId)
	_, err = rc.Do("set", redisKey, formattedUUID, "EX", exp)
	return formattedUUID, devId, err
}

// GetClientAuthInfoByDeviceCode 通过设备代码获取客户端授权信息
func GetClientAuthInfoByDeviceCode(deviceCode string) (*ClientAuthInfo, error) {
	rc := redisPkg.GetCoon()
	redisKey := fmt.Sprintf("copilot.proxy.map.%s", deviceCode)
	userCode, err := rc.Do("get", redisKey)
	if err != nil {
		return nil, err
	}
	redisKey = fmt.Sprintf("copilot.proxy.%s", userCode)
	authInfoData, err := redis.Bytes(rc.Do("get", redisKey))
	if err != nil {
		return nil, err
	}
	authInfo := &ClientAuthInfo{}
	err = json.Unmarshal(authInfoData, &authInfo)
	return authInfo, err
}

func GetClientAuthInfo(code string) (ClientAuthInfo, error) {
	rc := redisPkg.GetCoon()
	redisKey := fmt.Sprintf("copilot.proxy.%s", code)
	authInfoData, err := redis.Bytes(rc.Do("get", redisKey))
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
	for len(newUUID) < 40 {
		ud, _ := uuid.NewV4()
		newUUID += strings.Replace(ud.String(), "-", "", -1)
	}
	return newUUID[:codeLen]
}

// UpdateClientAuthStatusByDeviceCode 更新客户端授权码通过设备代码
func UpdateClientAuthStatusByDeviceCode(deviceCode string, cardCode string) error {
	rc := redisPkg.GetCoon()
	redisKey := fmt.Sprintf("copilot.proxy.map.%s", deviceCode)
	uCode, err := rc.Do("get", redisKey)
	if err != nil {
		return err
	}
	redisKey = fmt.Sprintf("copilot.proxy.%s", uCode)
	authInfoData, err := redis.Bytes(rc.Do("get", redisKey))
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
	_, err = rc.Do("set", redisKey, authInfoData)
	return err
}

func RemoveClientAuthInfoByDeviceCode(deviceCode string) error {
	rc := redisPkg.GetCoon()
	redisKey := fmt.Sprintf("copilot.proxy.map.%s", deviceCode)
	uCode, err := rc.Do("get", redisKey)
	if err != nil {
		return err
	}
	redisKey = fmt.Sprintf("copilot.proxy.%s", uCode)
	_, err = rc.Do("del", redisKey)
	if err != nil {
		return err
	}
	redisKey = fmt.Sprintf("copilot.proxy.map.%s", deviceCode)
	_, err = rc.Do("del", redisKey)
	return err
}
