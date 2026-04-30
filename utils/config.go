package utils

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

// LoadConfig config file
func LoadConfig() {
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found, using environment variables: %s\n", err.Error())
	}

	viper.AutomaticEnv()
	viper.SetDefault("DB_AUTO_MIGRATE", false)
	viper.SetDefault("DB_RUN_MIGRATIONS", false)
	viper.SetDefault("APP_PORT", 8080)

	for _, key := range []string{
		"AWS_REGION",
		"AWS_ACCESS_KEY_ID",
		"AWS_SECRET_ACCESS_KEY",
		"COGNITO_USER_POOL_ID",
		"COGNITO_CLIENT_ID",
		"PARTNER_COGNITO_USER_POOL_ID",
		"PARTNER_COGNITO_CLIENT_ID",
		"INTERNAL_API_KEY",
	} {
		if val := os.Getenv(key); val != "" {
			viper.Set(key, val)
		}
	}
}
