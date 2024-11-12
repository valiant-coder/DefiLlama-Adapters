package utils

import (
	"crypto/md5"
	"encoding/base32"
	"strconv"
	"time"

	"crypto/hmac"
	"crypto/sha1"
)

var googleAuthConfig = struct {
	secretSize int
	// 1 <= windowSize <= 17
	windowSize int8
}{
	secretSize: 15,
	windowSize: 1,
}

func GenerateSecretKey(uid string) (secret string, err error) {
	buffer := md5.Sum([]byte(uid))
	return base32.StdEncoding.EncodeToString(buffer[0:googleAuthConfig.secretSize]), nil
}

func VerifyGoogleAuth(secret string, verifyCode string) (bool, error) {
	c, err := strconv.ParseUint(verifyCode, 10, 64)
	if err != nil {
		return false, err
	}
	return CheckCode(secret, c, time.Now().UnixNano()/int64(time.Millisecond)), nil
}

func CheckCode(secret string, code uint64, timeMsec int64) bool {
	decodedSecret, err := base32.StdEncoding.DecodeString(secret)
	if err != nil {
		return false
	}
	t := (timeMsec / 1000) / 30
	for i := -googleAuthConfig.windowSize; i <= googleAuthConfig.windowSize; i++ {
		var hash uint64
		hash, _ = verifyCode(decodedSecret, t)
		if hash == code {
			return true
		}
	}
	return false
}

func verifyCode(key []byte, t int64) (uint64, error) {
	var data = [8]byte{}
	value := t
	for i := 8; i > 0; value >>= 8 {
		i--
		data[i] = byte(value)
	}
	mac := hmac.New(sha1.New, key)
	_, _ = mac.Write(data[:])
	hash := mac.Sum(nil)
	offset := hash[20-1] & 0xF
	truncatedHash := uint64(0)
	for i := 0; i < 4; i++ {
		truncatedHash <<= 8
		truncatedHash |= uint64(hash[int(offset)+i] & 0xFF)
	}
	truncatedHash &= 0x7FFFFFFF
	truncatedHash %= 1000000
	return truncatedHash, nil

}
