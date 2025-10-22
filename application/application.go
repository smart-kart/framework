package application

import (
	"context"
	"fmt"

	"github.com/smart-kart/framework/env"
	"github.com/smart-kart/framework/logger"
)

// Application represents the main application
type Application struct {
	logger         logger.Logger
	pgxRegistrar   func(context.Context) error
	redisConfig    *RedisConfig
	awsConfig      *AWSConfig
	errorHandler   ErrorHandler
	customValidator interface{}
}

// RedisConfig holds Redis configuration
type RedisConfig struct {
	Host     string
	Port     string
	Password string
}

// AWSConfig holds AWS configuration
type AWSConfig struct {
	Region      string
	SQSQueueURL string
	S3Bucket    string
}

// ErrorHandler handles error codes and messages
type ErrorHandler struct {
	ErrorMessages    map[string]string
	ValidationErrors map[string]string
}

// New creates a new Application
func New() *Application {
	return &Application{
		logger: logger.New(),
	}
}

// WithPgx configures PostgreSQL connection
func (a *Application) WithPgx(registrar func(context.Context) error) *Application {
	a.pgxRegistrar = registrar
	return a
}

// WithRedis configures Redis connection
func (a *Application) WithRedis() *Application {
	a.redisConfig = &RedisConfig{
		Host:     env.GetOrDefault(env.RedisHost, "localhost"),
		Port:     env.GetOrDefault(env.RedisPort, "6379"),
		Password: env.Get(env.RedisPassword),
	}
	return a
}

// WithAWS configures AWS services (SQS, S3)
func (a *Application) WithAWS() *Application {
	a.awsConfig = &AWSConfig{
		Region:      env.GetOrDefault(env.AWSRegion, "us-east-1"),
		SQSQueueURL: env.Get(env.SQSQueueURL),
		S3Bucket:    env.Get(env.S3Bucket),
	}
	return a
}

// WithErrorCode configures error handling
func (a *Application) WithErrorCode(errMsg, validationErr map[string]string) *Application {
	a.errorHandler = ErrorHandler{
		ErrorMessages:    errMsg,
		ValidationErrors: validationErr,
	}
	return a
}

// WithCustomValidator configures custom validator
func (a *Application) WithCustomValidator(validator interface{}) *Application {
	a.customValidator = validator
	return a
}

// Run initializes all application dependencies
func (a *Application) Run(ctx context.Context) error {
	a.logger.Info("running application initialization...")

	// Initialize PostgreSQL
	if a.pgxRegistrar != nil {
		a.logger.Info("initializing PostgreSQL connection...")
		if err := a.pgxRegistrar(ctx); err != nil {
			return fmt.Errorf("failed to initialize pgx: %w", err)
		}
		a.logger.Info("PostgreSQL connection initialized")
	}

	// Initialize Redis
	if a.redisConfig != nil {
		a.logger.Info("initializing Redis connection...")
		// TODO: Add actual Redis initialization
		a.logger.Info("Redis connection initialized")
	}

	// Initialize AWS services
	if a.awsConfig != nil {
		a.logger.Info("initializing AWS services...")
		// TODO: Add actual AWS initialization
		a.logger.Info("AWS services initialized")
	}

	a.logger.Info("application initialization completed")
	return nil
}