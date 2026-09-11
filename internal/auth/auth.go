package auth

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net/http"

	"github.com/golang-jwt/jwt/v5"
)

var ErrInvalidToken = errors.New("auth: invalid token")

// jwk is one entry from a JWKS ("JSON Web Key Set") response. Supabase
// signs tokens with ES256 (an asymmetric algorithm), so verification
// only needs the public half of the key pair — never a shared secret.
type jwk struct {
	Kid string `json:"kid"`
	Kty string `json:"kty"`
	Crv string `json:"crv"`
	X   string `json:"x"`
	Y   string `json:"y"`
}

type jwks struct {
	Keys []jwk `json:"keys"`
}

// Verifier holds the public keys fetched from a project's JWKS
// endpoint, keyed by kid ("key id") — a token's header names which
// key signed it, so multiple keys (e.g. during Supabase's own key
// rotation) can be supported at once.
type Verifier struct {
	keys map[string]*ecdsa.PublicKey
}

// NewVerifier fetches the JWKS once, at construction time, and parses
// each EC public key it contains. Fails fast (returns an error) if
// the endpoint is unreachable or returns nothing usable — same
// fail-fast-at-startup pattern as db.New/config.Load, since nothing
// can verify a single request without this.
func NewVerifier(jwksURL string) (*Verifier, error) {
	resp, err := http.Get(jwksURL)
	if err != nil {
		return nil, fmt.Errorf("%w: fetch jwks: %v", ErrInvalidToken, err)
	}
	defer resp.Body.Close()

	var set jwks
	if err := json.NewDecoder(resp.Body).Decode(&set); err != nil {
		return nil, fmt.Errorf("%w: decode jwks: %v", ErrInvalidToken, err)
	}

	keys := make(map[string]*ecdsa.PublicKey)
	for _, k := range set.Keys {
		// Only P-256 EC keys are supported — the only kind this
		// project's Supabase instance actually issues. A different
		// key type/curve is skipped rather than causing a hard
		// failure, so an unrelated key in the set can't break startup.
		if k.Kty != "EC" || k.Crv != "P-256" {
			continue
		}

		xBytes, err := base64.RawURLEncoding.DecodeString(k.X)
		if err != nil {
			return nil, fmt.Errorf("%w: decode key %s: %v", ErrInvalidToken, k.Kid, err)
		}
		yBytes, err := base64.RawURLEncoding.DecodeString(k.Y)
		if err != nil {
			return nil, fmt.Errorf("%w: decode key %s: %v", ErrInvalidToken, k.Kid, err)
		}

		keys[k.Kid] = &ecdsa.PublicKey{
			Curve: elliptic.P256(),
			X:     new(big.Int).SetBytes(xBytes),
			Y:     new(big.Int).SetBytes(yBytes),
		}
	}

	if len(keys) == 0 {
		return nil, fmt.Errorf("%w: no usable EC keys in jwks response", ErrInvalidToken)
	}

	return &Verifier{keys: keys}, nil
}

// VerifyToken parses and verifies a Supabase-issued JWT, returning the
// user's ID (the "sub" claim) if it's genuinely valid.
func (v *Verifier) VerifyToken(tokenString string) (string, error) {
	claims := &jwt.RegisteredClaims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
		// Algorithm-confusion guard, same purpose as the original
		// HS256 version's check — refuse to verify against whatever
		// algorithm a forged token claims, only ECDSA (the family
		// ES256 belongs to) is accepted.
		if _, ok := t.Method.(*jwt.SigningMethodECDSA); !ok {
			return nil, fmt.Errorf("%w: unexpected signing method %v", ErrInvalidToken, t.Header["alg"])
		}

		kid, ok := t.Header["kid"].(string)
		if !ok {
			return nil, fmt.Errorf("%w: missing kid header", ErrInvalidToken)
		}

		key, ok := v.keys[kid]
		if !ok {
			return nil, fmt.Errorf("%w: unknown key id %s", ErrInvalidToken, kid)
		}

		return key, nil
	})
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrInvalidToken, err)
	}
	if !token.Valid {
		return "", ErrInvalidToken
	}
	if claims.Subject == "" {
		return "", fmt.Errorf("%w: missing subject claim", ErrInvalidToken)
	}

	return claims.Subject, nil
}
