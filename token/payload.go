package token

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var (
	ErrInvalidToken = errors.New("token is invalid")
	ErrExpiredToken = errors.New("token has expired")
)

// Payload 保存项目里需要的自定义声明（claims）。
// 同时它实现了 jwt/v5 要求的一组声明 Getter。
type Payload struct {
	ID        uuid.UUID `json:"id"`
	Username  string    `json:"username"`
	IssuedAt  time.Time `json:"issued_at"`
	ExpiredAt time.Time `json:"expired_at"`
}

// NewPayload 创建一个新的载荷，包含随机 token ID 和过期时间。
func NewPayload(username string, duration time.Duration) (*Payload, error) {
	tokenID, err := uuid.NewRandom()
	if err != nil {
		return nil, err
	}

	payload := &Payload{
		ID:        tokenID,
		Username:  username,
		IssuedAt:  time.Now(),
		ExpiredAt: time.Now().Add(duration),
	}
	return payload, nil
}

// Validate 会在 jwt/v5 校验声明时被调用。
// 这里仅检查 token 是否过期。
func (payload *Payload) Validate() error {
	if time.Now().After(payload.ExpiredAt) {
		return ErrExpiredToken
	}
	return nil
}

// GetExpirationTime 返回 exp（过期时间）声明。
func (payload *Payload) GetExpirationTime() (*jwt.NumericDate, error) {
	return jwt.NewNumericDate(payload.ExpiredAt), nil
}

// GetIssuedAt 返回 iat（签发时间）声明。
func (payload *Payload) GetIssuedAt() (*jwt.NumericDate, error) {
	return jwt.NewNumericDate(payload.IssuedAt), nil
}

// GetNotBefore 返回 nbf（生效时间）声明。
// 这里使用签发时间作为开始生效时间。
func (payload *Payload) GetNotBefore() (*jwt.NumericDate, error) {
	return jwt.NewNumericDate(payload.IssuedAt), nil
}

// GetIssuer 返回 iss（签发者）声明。
func (payload *Payload) GetIssuer() (string, error) {
	return "", nil
}

// GetSubject 返回 sub（主题）声明。
func (payload *Payload) GetSubject() (string, error) {
	return payload.Username, nil
}

// GetAudience 返回 aud（受众）声明。
func (payload *Payload) GetAudience() (jwt.ClaimStrings, error) {
	return jwt.ClaimStrings{}, nil
}
