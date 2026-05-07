package utils

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func generateTestKey(t *testing.T) *rsa.PrivateKey {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	return key
}

func serveJWKS(t *testing.T, kid string, pub *rsa.PublicKey) *httptest.Server {
	t.Helper()
	nBytes := pub.N.Bytes()
	eBytes := big.NewInt(int64(pub.E)).Bytes()

	jwks := map[string]any{
		"keys": []map[string]any{
			{
				"kid": kid,
				"kty": "RSA",
				"alg": "RS256",
				"use": "sig",
				"n":   base64.RawURLEncoding.EncodeToString(nBytes),
				"e":   base64.RawURLEncoding.EncodeToString(eBytes),
			},
		},
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(jwks)
	}))
	t.Cleanup(srv.Close)
	return srv
}

func signJWT(t *testing.T, key *rsa.PrivateKey, kid string, claims map[string]any) string {
	t.Helper()
	header := map[string]string{"alg": "RS256", "kid": kid, "typ": "JWT"}
	hJSON, _ := json.Marshal(header)
	cJSON, _ := json.Marshal(claims)

	h := base64.RawURLEncoding.EncodeToString(hJSON)
	p := base64.RawURLEncoding.EncodeToString(cJSON)
	signingInput := h + "." + p

	hash := sha256.Sum256([]byte(signingInput))
	sig, err := rsa.SignPKCS1v15(rand.Reader, key, crypto.SHA256, hash[:])
	require.NoError(t, err)

	return signingInput + "." + base64.RawURLEncoding.EncodeToString(sig)
}

func testCognitoService(t *testing.T) (*CognitoService, *rsa.PrivateKey, string) {
	t.Helper()
	key := generateTestKey(t)
	kid := "test-kid-1"
	srv := serveJWKS(t, kid, &key.PublicKey)
	issuer := "https://cognito-idp.ap-northeast-1.amazonaws.com/test-pool"

	cs := &CognitoService{
		userPoolID: "test-pool",
		clientID:   "test-client",
		issuer:     issuer,
		jwks:       newJWKSCache(srv.URL),
	}
	return cs, key, kid
}

func validClaims(issuer, clientID string) map[string]any {
	return map[string]any{
		"sub":       "user-123",
		"email":     "test@example.com",
		"iss":       issuer,
		"aud":       clientID,
		"token_use": "id",
		"exp":       float64(time.Now().Add(time.Hour).Unix()),
	}
}

func TestVerifyIDToken_ValidToken(t *testing.T) {
	cs, key, kid := testCognitoService(t)
	token := signJWT(t, key, kid, validClaims(cs.issuer, cs.clientID))

	claims, err := cs.VerifyIDToken(token)

	assert.NoError(t, err)
	assert.Equal(t, "user-123", claims["sub"])
	assert.Equal(t, "test@example.com", claims["email"])
}

func TestVerifyIDToken_InvalidFormat(t *testing.T) {
	cs, _, _ := testCognitoService(t)
	_, err := cs.VerifyIDToken("not.a.valid.jwt.format")
	assert.Error(t, err)

	_, err = cs.VerifyIDToken("onlyone")
	assert.Error(t, err)
}

func TestVerifyIDToken_TamperedPayload(t *testing.T) {
	cs, key, kid := testCognitoService(t)
	token := signJWT(t, key, kid, validClaims(cs.issuer, cs.clientID))

	parts := splitJWT(token)
	tampered := map[string]any{"sub": "hacker", "email": "evil@bad.com", "iss": cs.issuer, "aud": cs.clientID, "token_use": "id", "exp": float64(time.Now().Add(time.Hour).Unix())}
	tJSON, _ := json.Marshal(tampered)
	parts[1] = base64.RawURLEncoding.EncodeToString(tJSON)
	forgedToken := parts[0] + "." + parts[1] + "." + parts[2]

	_, err := cs.VerifyIDToken(forgedToken)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "signature")
}

func TestVerifyIDToken_WrongSigningKey(t *testing.T) {
	cs, _, kid := testCognitoService(t)
	wrongKey := generateTestKey(t)

	token := signJWT(t, wrongKey, kid, validClaims(cs.issuer, cs.clientID))

	_, err := cs.VerifyIDToken(token)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "signature")
}

func TestVerifyIDToken_ExpiredToken(t *testing.T) {
	cs, key, kid := testCognitoService(t)
	claims := validClaims(cs.issuer, cs.clientID)
	claims["exp"] = float64(time.Now().Add(-time.Hour).Unix())

	token := signJWT(t, key, kid, claims)
	_, err := cs.VerifyIDToken(token)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "expired")
}

func TestVerifyIDToken_WrongIssuer(t *testing.T) {
	cs, key, kid := testCognitoService(t)
	claims := validClaims(cs.issuer, cs.clientID)
	claims["iss"] = "https://evil-issuer.example.com"

	token := signJWT(t, key, kid, claims)
	_, err := cs.VerifyIDToken(token)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "issuer")
}

func TestVerifyIDToken_WrongAudience(t *testing.T) {
	cs, key, kid := testCognitoService(t)
	claims := validClaims(cs.issuer, cs.clientID)
	claims["aud"] = "wrong-client-id"

	token := signJWT(t, key, kid, claims)
	_, err := cs.VerifyIDToken(token)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "audience")
}

func TestVerifyIDToken_NotIDToken(t *testing.T) {
	cs, key, kid := testCognitoService(t)
	claims := validClaims(cs.issuer, cs.clientID)
	claims["token_use"] = "access"

	token := signJWT(t, key, kid, claims)
	_, err := cs.VerifyIDToken(token)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not an ID token")
}

func TestVerifyIDToken_UnknownKid(t *testing.T) {
	cs, key, _ := testCognitoService(t)
	token := signJWT(t, key, "unknown-kid", validClaims(cs.issuer, cs.clientID))

	_, err := cs.VerifyIDToken(token)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "key")
}

func TestDecodeAdminToken_NilProvider(t *testing.T) {
	original := AdminCognitoProvider
	AdminCognitoProvider = nil
	defer func() { AdminCognitoProvider = original }()

	_, err := DecodeAdminToken("some.token.here")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not configured")
}

func TestDecodePartnerToken_NilProvider(t *testing.T) {
	original := PartnerCognitoProvider
	PartnerCognitoProvider = nil
	defer func() { PartnerCognitoProvider = original }()

	_, err := DecodePartnerToken("some.token.here")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not configured")
}

func TestJWKSCache_RefreshesOnExpiry(t *testing.T) {
	key := generateTestKey(t)
	kid := "kid-1"
	fetchCount := 0

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fetchCount++
		nBytes := key.PublicKey.N.Bytes()
		eBytes := big.NewInt(int64(key.PublicKey.E)).Bytes()
		json.NewEncoder(w).Encode(map[string]any{
			"keys": []map[string]any{{
				"kid": kid, "kty": "RSA",
				"n": base64.RawURLEncoding.EncodeToString(nBytes),
				"e": base64.RawURLEncoding.EncodeToString(eBytes),
			}},
		})
	}))
	defer srv.Close()

	cache := newJWKSCache(srv.URL)
	cache.ttl = 0 // force refresh every time

	_, err := cache.getKey(kid)
	assert.NoError(t, err)
	_, err = cache.getKey(kid)
	assert.NoError(t, err)
	assert.Equal(t, 2, fetchCount)
}

func splitJWT(token string) [3]string {
	var parts [3]string
	start := 0
	idx := 0
	for i, c := range token {
		if c == '.' {
			parts[idx] = token[start:i]
			start = i + 1
			idx++
		}
	}
	parts[idx] = token[start:]
	return parts
}

func TestVerifyIDToken_ForgedBase64Token_Rejected(t *testing.T) {
	cs, _, _ := testCognitoService(t)

	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"RS256","kid":"test-kid-1"}`))
	payload := base64.RawURLEncoding.EncodeToString([]byte(fmt.Sprintf(`{"sub":"hacker","iss":"%s","aud":"%s","exp":%d,"token_use":"id"}`, cs.issuer, cs.clientID, time.Now().Add(time.Hour).Unix())))
	fakeSignature := base64.RawURLEncoding.EncodeToString([]byte("not-a-real-signature"))
	forgedToken := header + "." + payload + "." + fakeSignature

	_, err := cs.VerifyIDToken(forgedToken)
	assert.Error(t, err)
}
