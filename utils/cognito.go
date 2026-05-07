package utils

import (
	"context"
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider/types"
	"github.com/spf13/viper"
)

type CognitoService struct {
	client     *cognitoidentityprovider.Client
	userPoolID string
	clientID   string
	issuer     string
	jwks       *jwksCache
}

type CognitoAuthResult struct {
	AccessToken   string
	IDToken       string
	RefreshToken  string
	ExpiresIn     int32
	ChallengeName string
	Session       string
}

var AdminCognitoProvider *CognitoService
var PartnerCognitoProvider *CognitoService

func InitCognito() error {
	cfg, err := loadAWSConfig(context.TODO())
	if err != nil {
		return fmt.Errorf("failed to load AWS config: %w", err)
	}

	region := strings.TrimSpace(viper.GetString("AWS_REGION"))

	adminPoolID := viper.GetString("COGNITO_USER_POOL_ID")
	adminClientID := viper.GetString("COGNITO_CLIENT_ID")

	if adminPoolID != "" && adminClientID != "" {
		issuer := fmt.Sprintf("https://cognito-idp.%s.amazonaws.com/%s", region, adminPoolID)
		AdminCognitoProvider = &CognitoService{
			client:     cognitoidentityprovider.NewFromConfig(cfg),
			userPoolID: adminPoolID,
			clientID:   adminClientID,
			issuer:     issuer,
			jwks:       newJWKSCache(issuer + "/.well-known/jwks.json"),
		}
		log.Println("Admin Cognito service initialized successfully")
	} else {
		log.Printf("Admin Cognito config incomplete - PoolID: %s, ClientID: %s",
			ifEmpty(adminPoolID, "empty"),
			ifEmpty(adminClientID, "empty"),
		)
	}

	partnerPoolID := viper.GetString("PARTNER_COGNITO_USER_POOL_ID")
	partnerClientID := viper.GetString("PARTNER_COGNITO_CLIENT_ID")

	if partnerPoolID != "" && partnerClientID != "" {
		issuer := fmt.Sprintf("https://cognito-idp.%s.amazonaws.com/%s", region, partnerPoolID)
		PartnerCognitoProvider = &CognitoService{
			client:     cognitoidentityprovider.NewFromConfig(cfg),
			userPoolID: partnerPoolID,
			clientID:   partnerClientID,
			issuer:     issuer,
			jwks:       newJWKSCache(issuer + "/.well-known/jwks.json"),
		}
		log.Println("Partner Cognito service initialized successfully")
	} else {
		log.Printf("Partner Cognito config incomplete - PoolID: %s, ClientID: %s",
			ifEmpty(partnerPoolID, "empty"),
			ifEmpty(partnerClientID, "empty"),
		)
	}

	if AdminCognitoProvider == nil && PartnerCognitoProvider == nil {
		log.Println("Warning: No Cognito providers initialized")
	}

	return nil
}

func ifEmpty(val, defaultVal string) string {
	if val == "" {
		return defaultVal
	}
	return "set"
}

func GetAdminCognitoService() *CognitoService {
	return AdminCognitoProvider
}

func GetPartnerCognitoService() *CognitoService {
	return PartnerCognitoProvider
}

// VerifyIDToken verifies the JWT signature against the Cognito JWKS and
// validates issuer, audience, expiration, and token_use claims.
func (cs *CognitoService) VerifyIDToken(tokenString string) (map[string]interface{}, error) {
	parts := strings.Split(tokenString, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid token format")
	}

	headerJSON, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return nil, fmt.Errorf("invalid token header: %w", err)
	}

	var header struct {
		Kid string `json:"kid"`
		Alg string `json:"alg"`
	}
	if err := json.Unmarshal(headerJSON, &header); err != nil {
		return nil, fmt.Errorf("invalid token header: %w", err)
	}

	if header.Alg != "RS256" {
		return nil, fmt.Errorf("unsupported algorithm: %s", header.Alg)
	}

	pubKey, err := cs.jwks.getKey(header.Kid)
	if err != nil {
		return nil, fmt.Errorf("key lookup failed: %w", err)
	}

	signingInput := parts[0] + "." + parts[1]
	signature, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		return nil, fmt.Errorf("invalid signature encoding: %w", err)
	}

	hash := sha256.Sum256([]byte(signingInput))
	if err := rsa.VerifyPKCS1v15(pubKey, crypto.SHA256, hash[:], signature); err != nil {
		return nil, fmt.Errorf("invalid token signature")
	}

	payloadJSON, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, fmt.Errorf("invalid token payload: %w", err)
	}

	var claims map[string]interface{}
	if err := json.Unmarshal(payloadJSON, &claims); err != nil {
		return nil, fmt.Errorf("invalid token claims: %w", err)
	}

	if iss, _ := claims["iss"].(string); iss != cs.issuer {
		return nil, fmt.Errorf("invalid issuer: %s", iss)
	}

	if aud, _ := claims["aud"].(string); aud != cs.clientID {
		return nil, fmt.Errorf("invalid audience: %s", aud)
	}

	if tokenUse, _ := claims["token_use"].(string); tokenUse != "id" {
		return nil, fmt.Errorf("not an ID token (token_use=%s)", tokenUse)
	}

	exp, ok := claims["exp"].(float64)
	if !ok {
		return nil, fmt.Errorf("missing expiration claim")
	}
	if time.Now().Unix() > int64(exp) {
		return nil, fmt.Errorf("token has expired")
	}

	return claims, nil
}

// DecodeAdminToken verifies a JWT from the Admin Cognito pool.
func DecodeAdminToken(idToken string) (map[string]interface{}, error) {
	if AdminCognitoProvider == nil {
		return nil, fmt.Errorf("admin cognito not configured")
	}
	return AdminCognitoProvider.VerifyIDToken(idToken)
}

// DecodePartnerToken verifies a JWT from the Partner Cognito pool.
func DecodePartnerToken(idToken string) (map[string]interface{}, error) {
	if PartnerCognitoProvider == nil {
		return nil, fmt.Errorf("partner cognito not configured")
	}
	return PartnerCognitoProvider.VerifyIDToken(idToken)
}

func (cs *CognitoService) AuthenticateUser(
	ctx context.Context,
	email, password string,
) (*CognitoAuthResult, error) {
	input := &cognitoidentityprovider.InitiateAuthInput{
		AuthFlow: types.AuthFlowTypeUserPasswordAuth,
		ClientId: aws.String(cs.clientID),
		AuthParameters: map[string]string{
			"USERNAME": email,
			"PASSWORD": password,
		},
	}

	result, err := cs.client.InitiateAuth(ctx, input)
	if err != nil {
		errStr := err.Error()
		if strings.Contains(errStr, "NotAuthorizedException") {
			return nil, fmt.Errorf("incorrect username or password")
		}
		if strings.Contains(errStr, "UserNotFoundException") {
			return nil, fmt.Errorf("user not found")
		}
		if strings.Contains(errStr, "UserNotConfirmedException") {
			return nil, fmt.Errorf("user account is not confirmed")
		}
		return nil, fmt.Errorf("authentication failed: %w", err)
	}

	if result.ChallengeName != "" {
		return &CognitoAuthResult{
			ChallengeName: string(result.ChallengeName),
			Session:       aws.ToString(result.Session),
		}, nil
	}

	if result.AuthenticationResult == nil {
		return nil, fmt.Errorf("authentication result is nil")
	}

	return &CognitoAuthResult{
		AccessToken:  aws.ToString(result.AuthenticationResult.AccessToken),
		IDToken:      aws.ToString(result.AuthenticationResult.IdToken),
		RefreshToken: aws.ToString(result.AuthenticationResult.RefreshToken),
		ExpiresIn:    result.AuthenticationResult.ExpiresIn,
	}, nil
}

func (cs *CognitoService) RespondToAuthChallenge(
	ctx context.Context,
	challengeName, session, code, username string,
) (*CognitoAuthResult, error) {
	input := &cognitoidentityprovider.RespondToAuthChallengeInput{
		ClientId:      aws.String(cs.clientID),
		ChallengeName: types.ChallengeNameType(challengeName),
		Session:       aws.String(session),
		ChallengeResponses: map[string]string{
			"SOFTWARE_TOKEN_MFA_CODE": code,
			"USERNAME":                username,
		},
	}

	result, err := cs.client.RespondToAuthChallenge(ctx, input)
	if err != nil {
		errStr := err.Error()
		if strings.Contains(errStr, "NotAuthorizedException") {
			return nil, fmt.Errorf("incorrect MFA code")
		}
		if strings.Contains(errStr, "CodeMismatchException") {
			return nil, fmt.Errorf("incorrect MFA code")
		}
		return nil, fmt.Errorf("MFA challenge failed: %w", err)
	}

	if result.AuthenticationResult == nil {
		return nil, fmt.Errorf("authentication result is nil")
	}

	return &CognitoAuthResult{
		AccessToken:  aws.ToString(result.AuthenticationResult.AccessToken),
		IDToken:      aws.ToString(result.AuthenticationResult.IdToken),
		RefreshToken: aws.ToString(result.AuthenticationResult.RefreshToken),
		ExpiresIn:    result.AuthenticationResult.ExpiresIn,
	}, nil
}

func (cs *CognitoService) GetUser(ctx context.Context, accessToken string) (*cognitoidentityprovider.GetUserOutput, error) {
	input := &cognitoidentityprovider.GetUserInput{
		AccessToken: aws.String(accessToken),
	}
	return cs.client.GetUser(ctx, input)
}

func (cs *CognitoService) RefreshToken(
	ctx context.Context,
	refreshToken string,
) (*CognitoAuthResult, error) {
	input := &cognitoidentityprovider.InitiateAuthInput{
		AuthFlow: types.AuthFlowTypeRefreshTokenAuth,
		ClientId: aws.String(cs.clientID),
		AuthParameters: map[string]string{
			"REFRESH_TOKEN": refreshToken,
		},
	}

	result, err := cs.client.InitiateAuth(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("cognito token refresh failed: %w", err)
	}

	if result.AuthenticationResult == nil {
		return nil, fmt.Errorf("authentication result is nil")
	}

	return &CognitoAuthResult{
		AccessToken:  aws.ToString(result.AuthenticationResult.AccessToken),
		IDToken:      aws.ToString(result.AuthenticationResult.IdToken),
		RefreshToken: refreshToken,
		ExpiresIn:    result.AuthenticationResult.ExpiresIn,
	}, nil
}
