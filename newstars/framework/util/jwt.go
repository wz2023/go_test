package util

import (
	"errors"
	"golang.org/x/sync/singleflight"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

var GVA_Singleflight = &singleflight.Group{}

var (
	TokenExpired     = errors.New("token is expired")
	TokenNotValidYet = errors.New("token not active yet")
	TokenMalformed   = errors.New("that's not even a token")
	TokenInvalid     = errors.New("couldn't handle this token")
)

type BaseClaims struct {
	UID      int    `json:"uid" form:"uid"`
	Platform int    `json:"platform" form:"platform"`
	PkgName  string `json:"pkg_name" form:"pkg_name"`
}

type CustomClaims struct {
	BaseClaims
	BufferTime int64
	jwt.RegisteredClaims
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type JWT struct {
	SigningKey  []byte
	BufferTime  string
	ExpiresTime string
	Issuer      string // 签名的发行者
}

func NewJWT() *JWT {
	return &JWT{}
}

func (j *JWT) SetSigningKey(signingKey []byte) *JWT {
	j.SigningKey = signingKey
	return j
}

func (j *JWT) SetBufferTime(bufferTime string) *JWT {
	j.BufferTime = bufferTime
	return j
}

func (j *JWT) SetExpiresTime(expiresTime string) *JWT {
	j.ExpiresTime = expiresTime
	return j
}

func (j *JWT) SetIssuer(issuer string) *JWT {
	j.Issuer = issuer
	return j
}

func (j *JWT) CreateClaims(baseClaims BaseClaims) CustomClaims {
	bf, _ := ParseDuration(j.BufferTime)
	ep, _ := ParseDuration(j.ExpiresTime)
	claims := CustomClaims{
		BaseClaims: baseClaims,
		BufferTime: int64(bf / time.Second), // 缓冲时间1天 缓冲时间内会获得新的token刷新令牌 此时一个用户会存在两个有效令牌 但是前端只留一个 另一个会丢失
		RegisteredClaims: jwt.RegisteredClaims{
			Audience:  jwt.ClaimStrings{"GVA"},                   // 受众
			NotBefore: jwt.NewNumericDate(time.Now().Add(-1000)), // 签名生效时间
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(ep)),    // 过期时间 7天  配置文件
			Issuer:    j.Issuer,                                  // 签名的发行者
		},
	}
	return claims
}

// CreateToken 创建一个token
func (j *JWT) CreateToken(claims CustomClaims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(j.SigningKey)
}

// CreateTokenByOldToken 旧token 换新token 使用归并回源避免并发问题
func (j *JWT) CreateTokenByOldToken(oldToken string, claims CustomClaims) (string, error) {
	v, err, _ := GVA_Singleflight.Do("JWT:"+oldToken, func() (interface{}, error) {
		return j.CreateToken(claims)
	})
	return v.(string), err
}

// ParseToken 解析 token
func (j *JWT) ParseToken(tokenString string) (*CustomClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (i interface{}, e error) {
		return j.SigningKey, nil
	})
	if err != nil {
		var ve *jwt.ValidationError
		if errors.As(err, &ve) {
			if ve.Errors&jwt.ValidationErrorMalformed != 0 {
				return nil, TokenMalformed
			} else if ve.Errors&jwt.ValidationErrorExpired != 0 {
				// Token is expired
				return nil, TokenExpired
			} else if ve.Errors&jwt.ValidationErrorNotValidYet != 0 {
				return nil, TokenNotValidYet
			} else {
				return nil, TokenInvalid
			}
		}
	}
	if token != nil {
		if claims, ok := token.Claims.(*CustomClaims); ok && token.Valid {
			return claims, nil
		}
		return nil, TokenInvalid

	} else {
		return nil, TokenInvalid
	}
}
