package jwks

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

const (
	defaultIssuer = "go-ojol-auth"
	p256CoordSize = 32
	defaultTTL    = 5 * time.Minute
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

type Claims struct {
	UserID    string
	Email     string
	Role      string
	SessionID string
}

type Verifier interface {
	Verify(token string) (*Claims, error)
}

type verifier struct {
	jwksURL    string
	issuer     string
	httpClient *http.Client
	ttl        time.Duration

	mu        sync.RWMutex
	keys      map[string]*ecdsa.PublicKey
	fetchedAt time.Time
}

func NewVerifierFromEnv() (Verifier, error) {
	jwksURL := os.Getenv("AUTH_JWKS_URL")
	if jwksURL == "" {
		return nil, errors.New("AUTH_JWKS_URL is required")
	}

	issuer := os.Getenv("JWT_ISSUER")
	if issuer == "" {
		issuer = defaultIssuer
	}

	return NewVerifier(jwksURL, issuer), nil
}

func NewVerifier(jwksURL, issuer string) Verifier {
	return &verifier{
		jwksURL: jwksURL,
		issuer:  issuer,
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
		ttl:  defaultTTL,
		keys: make(map[string]*ecdsa.PublicKey),
	}
}

func (v *verifier) Verify(raw string) (*Claims, error) {
	parser := jwt.Parser{
		ValidMethods: []string{jwt.SigningMethodES256.Alg()},
	}

	token, err := parser.Parse(raw, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodECDSA); !ok {
			return nil, fmt.Errorf("unexpected signing method %v", t.Header["alg"])
		}

		kid, _ := t.Header["kid"].(string)
		if kid == "" {
			return nil, errors.New("missing kid")
		}

		return v.keyForKID(kid)
	})
	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	mapClaims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("invalid token claims")
	}

	if v.issuer != "" && !mapClaims.VerifyIssuer(v.issuer, true) {
		return nil, errors.New("invalid issuer")
	}

	userID, err := claimString(mapClaims, "user_id")
	if err != nil {
		return nil, err
	}
	email, err := claimString(mapClaims, "email")
	if err != nil {
		return nil, err
	}
	role, err := claimString(mapClaims, "role")
	if err != nil {
		return nil, err
	}
	sessionID, err := claimString(mapClaims, "session_id")
	if err != nil {
		return nil, err
	}

	return &Claims{
		UserID:    userID,
		Email:     email,
		Role:      role,
		SessionID: sessionID,
	}, nil
}

func (v *verifier) keyForKID(kid string) (*ecdsa.PublicKey, error) {
	v.mu.RLock()
	fresh := !v.fetchedAt.IsZero() && time.Since(v.fetchedAt) < v.ttl
	key, ok := v.keys[kid]
	v.mu.RUnlock()
	if fresh && ok {
		return key, nil
	}

	if err := v.refresh(); err != nil {
		return nil, err
	}

	v.mu.RLock()
	key, ok = v.keys[kid]
	v.mu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("unknown kid %q", kid)
	}

	return key, nil
}

func (v *verifier) refresh() error {
	req, err := http.NewRequest(http.MethodGet, v.jwksURL, nil)
	if err != nil {
		return fmt.Errorf("jwks request: %w", err)
	}

	resp, err := v.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("jwks fetch: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("jwks fetch: status %d", resp.StatusCode)
	}

	var jwks JWKS
	if err := json.NewDecoder(resp.Body).Decode(&jwks); err != nil {
		return fmt.Errorf("jwks decode: %w", err)
	}

	keys := make(map[string]*ecdsa.PublicKey, len(jwks.Keys))
	for _, jwk := range jwks.Keys {
		if jwk.Kid == "" {
			continue
		}
		pub, err := publicKeyFromJWK(jwk)
		if err != nil {
			return fmt.Errorf("jwks key %q: %w", jwk.Kid, err)
		}
		keys[jwk.Kid] = pub
	}

	v.mu.Lock()
	v.keys = keys
	v.fetchedAt = time.Now()
	v.mu.Unlock()

	return nil
}

func publicKeyFromJWK(jwk JWK) (*ecdsa.PublicKey, error) {
	if jwk.Kty != "EC" || jwk.Crv != "P-256" {
		return nil, fmt.Errorf("unsupported key type %s/%s", jwk.Kty, jwk.Crv)
	}
	if jwk.Alg != "" && jwk.Alg != jwt.SigningMethodES256.Alg() {
		return nil, fmt.Errorf("unsupported alg %s", jwk.Alg)
	}

	xBytes, err := base64.RawURLEncoding.DecodeString(jwk.X)
	if err != nil {
		return nil, fmt.Errorf("decode x: %w", err)
	}
	yBytes, err := base64.RawURLEncoding.DecodeString(jwk.Y)
	if err != nil {
		return nil, fmt.Errorf("decode y: %w", err)
	}
	if len(xBytes) > p256CoordSize || len(yBytes) > p256CoordSize {
		return nil, errors.New("invalid P-256 coordinate size")
	}

	pub := &ecdsa.PublicKey{
		Curve: elliptic.P256(),
		X:     new(big.Int).SetBytes(xBytes),
		Y:     new(big.Int).SetBytes(yBytes),
	}
	if !pub.Curve.IsOnCurve(pub.X, pub.Y) {
		return nil, errors.New("public key is not on P-256")
	}

	return pub, nil
}

func claimString(claims jwt.MapClaims, key string) (string, error) {
	value, exists := claims[key]
	if !exists || value == nil {
		return "", fmt.Errorf("%s not found in token", key)
	}

	s, ok := value.(string)
	if !ok || s == "" {
		return "", fmt.Errorf("%s not found in token", key)
	}

	return s, nil
}
