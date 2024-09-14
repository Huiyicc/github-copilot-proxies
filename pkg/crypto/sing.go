package crypto

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"math/rand"
	"reflect"
	"sort"
	"strconv"
	"time"
)

// GetSign get the sign info
func GetSign(data interface{}, appSecret string) string {
	md5ctx := md5.New()
	switch v := reflect.ValueOf(data); v.Kind() {
	case reflect.String:
		md5ctx.Write([]byte(v.String() + appSecret))
		return hex.EncodeToString(md5ctx.Sum(nil))
	case reflect.Struct:
		orderStr := StructToMapSing(v.Interface(), appSecret)
		md5ctx.Write([]byte(orderStr))
		return hex.EncodeToString(md5ctx.Sum(nil))
	case reflect.Ptr:
		originType := v.Elem().Type()
		if originType.Kind() != reflect.Struct {
			return ""
		}
		dataType := reflect.TypeOf(data).Elem()
		dataVal := v.Elem()
		orderStr := buildOrderStr(dataType, dataVal, appSecret)
		md5ctx.Write([]byte(orderStr))
		return hex.EncodeToString(md5ctx.Sum(nil))
	default:
		return ""
	}
}

func buildOrderStr(t reflect.Type, v reflect.Value, appSecret string) (returnStr string) {
	keys := make([]string, 0, t.NumField())
	var data = make(map[string]interface{})
	for i := 0; i < t.NumField(); i++ {
		if t.Field(i).Tag.Get("json") == "sign" {
			continue
		}
		data[t.Field(i).Tag.Get("json")] = v.Field(i).Interface()

		keys = append(keys, t.Field(i).Tag.Get("json"))
	}
	sort.Sort(sort.StringSlice(keys))
	var buf bytes.Buffer
	for _, k := range keys {
		if data[k] == "" {
			continue
		}
		if buf.Len() > 0 {
			buf.WriteByte('&')
		}

		buf.WriteString(k)
		buf.WriteByte('=')
		switch vv := data[k].(type) {
		case string:
			buf.WriteString(vv)
		case int:
		case int8:
		case int16:
		case int32:
		case int64:
			buf.WriteString(strconv.FormatInt(int64(vv), 10))
		default:
			continue
		}
	}
	buf.WriteString("&secret=" + appSecret)
	returnStr = buf.String()
	return returnStr
}

func StructToMapSing(content interface{}, appSecret string) (returnStr string) {
	t := reflect.TypeOf(content)
	v := reflect.ValueOf(content)
	returnStr = buildOrderStr(t, v, appSecret)
	return returnStr
}

func EnSign(query, body, key string) string {
	//加密算法
	//r：随机数
	//t：时间戳
	//q：query参数md5
	//b：Body参数md5
	//k：密钥
	//组合成 k=%s&r=%d&t=%d&q=%s&b=%s
	//进行md5
	//最后组合成 r,t,md5
	rand.Seed(time.Now().Unix())
	r := rand.Intn(800000) + 100000
	t := time.Now().Unix()
	str := fmt.Sprintf("k=%s&r=%d&t=%d&q=%s&b=%s", key, r, t, GetMd5(query), GetMd5(body))
	return fmt.Sprintf("%d,%d,%s", r, t, GetSign(str, key))
}
