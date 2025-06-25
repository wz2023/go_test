package aescbc

// AuthInfo 认证参数
type AuthInfo struct {
	UUID         string
	SerialNumber string
	Mac          string
	IP           uint64
	EndTime      int64
}
