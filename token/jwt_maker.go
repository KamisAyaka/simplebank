package token

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const minSecretKeySize = 32

type JWTMaker struct {
	secretKey string
}

// NewJWTMaker 创建 JWT 生成器，使用 HMAC 密钥。
// 为了安全性，HS256 需要足够长度的密钥。
func NewJWTMaker(secretKey string) (Maker, error) {
	if len(secretKey) < minSecretKeySize {
		return nil, fmt.Errorf("invalid key size: must be at least %d characters", minSecretKeySize)
	}
	return &JWTMaker{secretKey}, nil
}

// CreateToken 使用自定义 Payload 作为 claims 生成并签名 JWT。
func (maker *JWTMaker) CreateToken(username string, duration time.Duration) (string, *Payload, error) {
	payload, err := NewPayload(username, duration)
	if err != nil {
		return "", payload, err
	}
	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, payload)
	token, err := jwtToken.SignedString([]byte(maker.secretKey))
	return token, payload, err
}

// VerifyToken 解析并校验 JWT。
// 过程包含：签名算法检查、签名校验、声明校验，最后返回 Payload。
func (maker *JWTMaker) VerifyToken(token string) (*Payload, error) {
	// keyFunc 会被 jwt 解析器调用，用于提供验签密钥。
	keyFunc := func(token *jwt.Token) (interface{}, error) {
		_, ok := token.Method.(*jwt.SigningMethodHMAC)
		if !ok {
			return nil, ErrInvalidToken
		}
		return []byte(maker.secretKey), nil
	}

	// 直接解析到我们自己的 claims 类型（Payload）。
	jwtToken, err := jwt.ParseWithClaims(token, &Payload{}, keyFunc)
	if err != nil {
		// jwt/v5 会包装底层 claims 错误，这里统一映射成项目内的业务错误。
		if errors.Is(err, ErrExpiredToken) || errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, ErrInvalidToken
	}
	if !jwtToken.Valid {
		return nil, ErrInvalidToken
	}

	// 从通用 Claims 做类型断言，拿到具体 Payload。
	payload, ok := jwtToken.Claims.(*Payload)
	if !ok {
		return nil, ErrInvalidToken
	}

	return payload, nil
}
