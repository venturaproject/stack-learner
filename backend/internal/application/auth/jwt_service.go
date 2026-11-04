package auth

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"boilerplate-learner/internal/shared/apperrors"
)

type JWTConfig struct {
	Secret         string
	AccessTTLSecs  int
	RefreshTTLSecs int
}

type Claims struct {
	jwt.RegisteredClaims
	Sub       int32  `json:"sub"`
	TokenType string `json:"token_type"`
}

type TokenPair struct {
	AccessToken      string
	RefreshToken     string
	RefreshJTI       string
	RefreshExpiresAt time.Time
}

type JWTService struct {
	cfg JWTConfig
}

func NewJWTService(cfg JWTConfig) *JWTService {
	return &JWTService{cfg: cfg}
}

func (s *JWTService) GenerateTokens(userID int32) (*TokenPair, error) {
	now := time.Now()
	accessExp := now.Add(time.Duration(s.cfg.AccessTTLSecs) * time.Second)
	refreshExp := now.Add(time.Duration(s.cfg.RefreshTTLSecs) * time.Second)
	refreshJTI := uuid.New().String()

	accessToken, err := s.signToken(userID, uuid.New().String(), "access", accessExp)
	if err != nil {
		return nil, err
	}
	refreshToken, err := s.signToken(userID, refreshJTI, "refresh", refreshExp)
	if err != nil {
		return nil, err
	}
	return &TokenPair{
		AccessToken:      accessToken,
		RefreshToken:     refreshToken,
		RefreshJTI:       refreshJTI,
		RefreshExpiresAt: refreshExp,
	}, nil
}

func (s *JWTService) ParseAccessToken(tokenStr string) (*Claims, error) {
	return s.parseToken(tokenStr, "access")
}

func (s *JWTService) ParseRefreshToken(tokenStr string) (*Claims, error) {
	return s.parseToken(tokenStr, "refresh")
}

func (s *JWTService) signToken(userID int32, jti, tokenType string, exp time.Time) (string, error) {
	claims := Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(exp),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ID:        jti,
		},
		Sub:       userID,
		TokenType: tokenType,
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(s.cfg.Secret))
}

func (s *JWTService) parseToken(tokenStr, expectedType string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, apperrors.Unauthorized("Algoritmo inválido")
		}
		return []byte(s.cfg.Secret), nil
	})
	if err != nil || !token.Valid {
		return nil, apperrors.Unauthorized("Token inválido o expirado")
	}
	claims, ok := token.Claims.(*Claims)
	if !ok || claims.TokenType != expectedType {
		return nil, apperrors.Unauthorized("Tipo de token incorrecto")
	}
	return claims, nil
}
