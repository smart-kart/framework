package env

import (
	"bufio"
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

// LoadFromEnv loads environment variables from a .env file
func LoadFromEnv(filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		// If .env file doesn't exist, that's okay - just skip it
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("failed to open .env file: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Parse KEY=VALUE format
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			return fmt.Errorf("invalid format at line %d: %s", lineNum, line)
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// Remove quotes if present
		value = strings.Trim(value, `"'`)

		// Only set if not already set in environment
		if os.Getenv(key) == "" {
			if err := os.Setenv(key, value); err != nil {
				return fmt.Errorf("failed to set env var %s: %w", key, err)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading .env file: %w", err)
	}

	return nil
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

// Validate checks that all required environment variables are set
func Validate(requiredVars ...string) error {
	var missing []string
	for _, key := range requiredVars {
		if os.Getenv(key) == "" {
			missing = append(missing, key)
		}
	}

	if len(missing) > 0 {
		return fmt.Errorf("missing required environment variables: %s", strings.Join(missing, ", "))
	}

	return nil
}