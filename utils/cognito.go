package utils

import (
	"context"
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
}

type CognitoAuthResult struct {
	AccessToken   string
	IDToken       string
	RefreshToken  string
	ExpiresIn     int32
	ChallengeName string
	Session       string
}

// Dual Cognito pools: admin and partner
var AdminCognitoProvider *CognitoService
var PartnerCognitoProvider *CognitoService

func InitCognito() error {
	cfg, err := loadAWSConfig(context.TODO())
	if err != nil {
		return fmt.Errorf("failed to load AWS config: %w", err)
	}

	// Initialize Admin Cognito pool
	adminPoolID := viper.GetString("COGNITO_USER_POOL_ID")
	adminClientID := viper.GetString("COGNITO_CLIENT_ID")

	if adminPoolID != "" && adminClientID != "" {
		AdminCognitoProvider = &CognitoService{
			client:     cognitoidentityprovider.NewFromConfig(cfg),
			userPoolID: adminPoolID,
			clientID:   adminClientID,
		}
		log.Println("Admin Cognito service initialized successfully")
	} else {
		log.Printf("Admin Cognito config incomplete - PoolID: %s, ClientID: %s",
			ifEmpty(adminPoolID, "empty"),
			ifEmpty(adminClientID, "empty"),
		)
	}

	// Initialize Partner Cognito pool
	partnerPoolID := viper.GetString("PARTNER_COGNITO_USER_POOL_ID")
	partnerClientID := viper.GetString("PARTNER_COGNITO_CLIENT_ID")

	if partnerPoolID != "" && partnerClientID != "" {
		PartnerCognitoProvider = &CognitoService{
			client:     cognitoidentityprovider.NewFromConfig(cfg),
			userPoolID: partnerPoolID,
			clientID:   partnerClientID,
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

// DecodeAdminToken decodes and validates a JWT from the Admin Cognito pool
func DecodeAdminToken(idToken string) (map[string]interface{}, error) {
	return decodeCognitoIDToken(idToken)
}

// DecodePartnerToken decodes and validates a JWT from the Partner Cognito pool
func DecodePartnerToken(idToken string) (map[string]interface{}, error) {
	return decodeCognitoIDToken(idToken)
}

func decodeCognitoIDToken(idToken string) (map[string]interface{}, error) {
	parts := strings.Split(idToken, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid token format")
	}

	payload := parts[1]
	decoded, err := base64.RawURLEncoding.DecodeString(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to decode token: %w", err)
	}

	var claims map[string]interface{}
	if err := json.Unmarshal(decoded, &claims); err != nil {
		return nil, fmt.Errorf("failed to unmarshal claims: %w", err)
	}

	// Validate token expiration
	if exp, ok := claims["exp"].(float64); ok {
		expirationTime := int64(exp)
		currentTime := GetCurrentTimestamp()

		if currentTime > expirationTime {
			log.Printf("Token expired: current=%d, exp=%d, diff=%d seconds",
				currentTime, expirationTime, currentTime-expirationTime)
			return nil, fmt.Errorf("token has expired")
		}
	} else {
		return nil, fmt.Errorf("token missing expiration (exp) claim")
	}

	return claims, nil
}

// GetCurrentTimestamp returns the current Unix timestamp in seconds
func GetCurrentTimestamp() int64 {
	return time.Now().Unix()
}
