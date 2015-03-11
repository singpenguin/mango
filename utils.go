package mango

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"strconv"
	"time"
)

func Timestamp() int64 {
	return time.Now().Unix()
}

func FormatInt(i int64, base ...interface{}) string {
	var b int
	if len(base) == 1 {
		b = base[0].(int)
	} else {
		b = 10
	}
	return strconv.FormatInt(i, b)
}

func ParseInt(s string) (int64, error) {
	return strconv.ParseInt(s, 10, 64)
}

func Base64Encode(src []byte) string {
	rs := base64.StdEncoding.EncodeToString(src)
	return rs
}

func Base64Decode(src string) (string, error) {
	rs, err := base64.StdEncoding.DecodeString(src)
	return string(rs), err
}

func HmacSha256(message string, key string) string {
	mac := hmac.New(sha256.New, []byte(key))
	mac.Write([]byte(message))
	exmac := mac.Sum(nil)
	return fmt.Sprintf("%x", exmac)
}

func unpack(s []string, vars ...*string) {
	for i, str := range s {
		*vars[i] = str
	}
}
