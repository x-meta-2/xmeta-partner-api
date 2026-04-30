package utils

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/spf13/viper"
)

func loadAWSConfig(ctx context.Context) (aws.Config, error) {
	region := strings.TrimSpace(viper.GetString("AWS_REGION"))
	if region == "" {
		region = strings.TrimSpace(os.Getenv("AWS_DEFAULT_REGION"))
	}
	if region == "" {
		return aws.Config{}, fmt.Errorf("aws region is not configured")
	}

	accessKeyID := strings.TrimSpace(viper.GetString("AWS_ACCESS_KEY_ID"))
	secretAccessKey := strings.TrimSpace(viper.GetString("AWS_SECRET_ACCESS_KEY"))

	if accessKeyID != "" || secretAccessKey != "" {
		if accessKeyID == "" || secretAccessKey == "" {
			return aws.Config{}, fmt.Errorf("aws static credentials are incomplete")
		}
		return awsconfig.LoadDefaultConfig(
			ctx,
			awsconfig.WithRegion(region),
			awsconfig.WithCredentialsProvider(
				credentials.NewStaticCredentialsProvider(accessKeyID, secretAccessKey, ""),
			),
		)
	}

	return awsconfig.LoadDefaultConfig(
		ctx,
		awsconfig.WithRegion(region),
	)
}
