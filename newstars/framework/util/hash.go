package util

import (
	"crypto/md5"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"math/rand"
	"time"
)

// Md5Hash 生成32位MD5
func Md5Hash(text string) string {
	ctx := md5.New()
	ctx.Write([]byte(text))
	return hex.EncodeToString(ctx.Sum(nil))
}

// RandomString 生成随机字符串
func RandomString(l int) string {
	bytes := []byte("0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	result := []byte{}
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < l; i++ {
		result = append(result, bytes[r.Intn(len(bytes))])
	}
	return string(result)
}

// RandomCode 生成随机码
func RandomCode(l int) string {
	bytes := []byte("0123456789")
	result := []byte{}
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < l; i++ {
		result = append(result, bytes[r.Intn(len(bytes))])
	}
	return string(result)
}

// EncryptPassword 生成密码串
func EncryptPassword(password, salt string) string {
	return Md5Hash(Md5Hash(password) + Md5Hash(salt))
}

func StringToIntHash(s string) uint64 {
	// 生成SHA256哈希
	hash := sha256.Sum256([]byte(s))

	// 取哈希的前8字节转换为uint64
	return binary.BigEndian.Uint64(hash[:8])
}
