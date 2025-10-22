package env

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// Environment variable keys
const (
	Service        = "SERVICE_NAME"
	Environment    = "ENVIRONMENT"
	ServerPort     = "SERVER_PORT"
	GRPCPort       = "GRPC_PORT"
	DBHost         = "DB_HOST"
	DBPort         = "DB_PORT"
	DBUser         = "DB_USER"
	DBPassword     = "DB_PASSWORD"
	DBName         = "DB_NAME"
	DBDrivers      = "DB_DRIVERS"
	RedisHost      = "REDIS_HOST"
	RedisPort      = "REDIS_PORT"
	RedisPassword  = "REDIS_PASSWORD"
	AWSRegion           = "AWS_REGION"
	SQSQueueURL         = "SQS_QUEUE_URL"
	S3Bucket            = "S3_BUCKET"
	JWTSecretKey        = "JWT_SECRET_KEY"
	JWTAccessTokenTTL   = "JWT_ACCESS_TOKEN_TTL"
	JWTRefreshTokenTTL  = "JWT_REFRESH_TOKEN_TTL"
	JWTIssuer           = "JWT_ISSUER"
)

// Environment types
const (
	UnitTest = "unittest"
	Dev      = "dev"
	Staging  = "staging"
	Prod     = "prod"
)

// Get retrieves an environment variable value
func Get(key string) string {
	return os.Getenv(key)
}

// GetOrDefault retrieves an environment variable or returns default
func GetOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// Set sets an environment variable
func Set(key, value string) error {
	return os.Setenv(key, value)
}

// GetList retrieves a comma-separated environment variable as a slice
func GetList(key string) []string {
	value := os.Getenv(key)
	if value == "" {
		return []string{}
	}
	parts := strings.Split(value, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

// LoadFromYAML loads environment variables from a YAML file
func LoadFromYAML(filePath string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read YAML file: %w", err)
	}

	var config map[string]interface{}
	if err := yaml.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("failed to parse YAML: %w", err)
	}

	for key, value := range config {
		if strValue, ok := value.(string); ok {
			if err := os.Setenv(key, strValue); err != nil {
				return fmt.Errorf("failed to set env var %s: %w", key, err)
			}
		}
	}

	return nil
}