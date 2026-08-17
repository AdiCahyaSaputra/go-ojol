package service

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

const (
	defaultIssuer = "go-ojol-auth"
	p256CoordSize = 32
)

type JWK struct {
	Kty string `json:"kty"`
	Crv string `json:"crv"`
	X   string `json:"x"`
	Y   string `json:"y"`
	Use string `json:"use"`
	Alg string `json:"alg"`
	Kid string `json:"kid"`
}

type JWKS struct {
	Keys []JWK `json:"keys"`
}

type JWTService interface {
	GenerateAccessToken(userId string, email string, role string) (string, error)
	GenerateRefreshToken() (string, time.Time)
	ValidateToken(token string) (*jwt.Token, error)
	GetUserIDByToken(token string) (string, error)
	GetEmailByToken(token string) (string, error)
	JWKS() JWKS
}

type jwtCustomClaim struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

type jwtService struct {
	privateKey    *ecdsa.PrivateKey
	publicKey     *ecdsa.PublicKey
	kid           string
	issuer        string
	jwks          JWKS
	accessExpiry  time.Duration
	refreshExpiry time.Duration
}

func NewJWTService() (JWTService, error) {
	path := os.Getenv("JWT_PRIVATE_KEY_PATH")
	if path == "" {
		return nil, errors.New("JWT_PRIVATE_KEY_PATH is required")
	}

	pemBytes, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read JWT private key: %w", err)
	}

	key, err := parseECPrivateKey(pemBytes)
	if err != nil {
		return nil, err
	}

	return NewJWTServiceFromKey(key, os.Getenv("JWT_KID"), os.Getenv("JWT_ISSUER"))
}

func NewJWTServiceFromKey(key *ecdsa.PrivateKey, kid, issuer string) (JWTService, error) {
	if key == nil {
		return nil, errors.New("private key is required")
	}
	if key.Curve != elliptic.P256() {
		return nil, errors.New("JWT private key must be P-256")
	}

	jwk, err := publicJWK(&key.PublicKey)
	if err != nil {
		return nil, err
	}

	if kid == "" {
		kid, err = jwkThumbprint(jwk)
		if err != nil {
			return nil, err
		}
	}
	jwk.Kid = kid

	if issuer == "" {
		issuer = defaultIssuer
	}

	return &jwtService{
		privateKey:    key,
		publicKey:     &key.PublicKey,
		kid:           kid,
		issuer:        issuer,
		jwks:          JWKS{Keys: []JWK{jwk}},
		accessExpiry:  time.Minute * 15,
		refreshExpiry: time.Hour * 24 * 7,
	}, nil
}

func parseECPrivateKey(pemBytes []byte) (*ecdsa.PrivateKey, error) {
	block, _ := pem.Decode(pemBytes)
	if block == nil {
		return nil, errors.New("invalid PEM: no block found")
	}

	if key, err := x509.ParseECPrivateKey(block.Bytes); err == nil {
		return key, nil
	}

	parsed, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("parse EC private key: %w", err)
	}

	key, ok := parsed.(*ecdsa.PrivateKey)
	if !ok {
		return nil, errors.New("JWT private key is not ECDSA")
	}
	return key, nil
}

func publicJWK(pub *ecdsa.PublicKey) (JWK, error) {
	if pub == nil || pub.Curve != elliptic.P256() {
		return JWK{}, errors.New("public key must be P-256")
	}

	return JWK{
		Kty: "EC",
		Crv: "P-256",
		X:   base64.RawURLEncoding.EncodeToString(pub.X.FillBytes(make([]byte, p256CoordSize))),
		Y:   base64.RawURLEncoding.EncodeToString(pub.Y.FillBytes(make([]byte, p256CoordSize))),
		Use: "sig",
		Alg: "ES256",
	}, nil
}

func jwkThumbprint(jwk JWK) (string, error) {
	canonical := fmt.Sprintf(`{"crv":%q,"kty":%q,"x":%q,"y":%q}`, jwk.Crv, jwk.Kty, jwk.X, jwk.Y)
	sum := sha256.Sum256([]byte(canonical))
	return base64.RawURLEncoding.EncodeToString(sum[:]), nil
}

func (j *jwtService) GenerateAccessToken(userId string, email string, role string) (string, error) {
	claims := jwtCustomClaim{
		UserID: userId,
		Email:  email,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(j.accessExpiry)),
			Issuer:    j.issuer,
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodES256, claims)
	token.Header["kid"] = j.kid
	return token.SignedString(j.privateKey)
}

func (j *jwtService) GenerateRefreshToken() (string, time.Time) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", time.Time{}
	}

	refreshToken := base64.StdEncoding.EncodeToString(b)
	expiresAt := time.Now().Add(j.refreshExpiry)

	return refreshToken, expiresAt
}

func (j *jwtService) parseToken(t_ *jwt.Token) (any, error) {
	if _, ok := t_.Method.(*jwt.SigningMethodECDSA); !ok {
		return nil, fmt.Errorf("unexpected signing method %v", t_.Header["alg"])
	}
	return j.publicKey, nil
}

func (j *jwtService) ValidateToken(token string) (*jwt.Token, error) {
	return jwt.Parse(token, j.parseToken)
}

func (j *jwtService) GetUserIDByToken(token string) (string, error) {
	return j.claimByToken(token, "user_id")
}

func (j *jwtService) GetEmailByToken(token string) (string, error) {
	return j.claimByToken(token, "email")
}

func (j *jwtService) claimByToken(token, key string) (string, error) {
	tToken, err := j.ValidateToken(token)
	if err != nil {
		return "", err
	}

	claims, ok := tToken.Claims.(jwt.MapClaims)
	if !ok {
		return "", errors.New("invalid token claims")
	}

	value, exists := claims[key]
	if !exists || value == nil {
		return "", fmt.Errorf("%s not found in token", key)
	}

	s := fmt.Sprintf("%v", value)
	if s == "" || s == "<nil>" {
		return "", fmt.Errorf("%s not found in token", key)
	}

	return s, nil
}

func (j *jwtService) JWKS() JWKS {
	return j.jwks
}
