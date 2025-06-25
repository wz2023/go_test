package cachekey

import "fmt"

func GetUserKey(id string) string {
	return fmt.Sprintf("user:%s", id)
}

func GetUserInfoKey(id string) string {
	return fmt.Sprintf("userinfo:%s", id)
}
