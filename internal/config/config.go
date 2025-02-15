package config

import (
	"context"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
)

const (
	defaultPort   = "9898"
	defaultRegion = "us-east-1"
)

type Config struct {
	Port               string
	AWSAccessKey       string
	AWSSecretAccessKey string
}

func New() *Config {
	return &Config{
		Port:               defaultPort,
		AWSAccessKey:       os.Getenv("AWS_ACCESS_KEY"),
		AWSSecretAccessKey: os.Getenv("AWS_SECRET_ACCESS_KEY"),
	}
}

func (c *Config) LoadAWSConfig() (aws.Config, error) {
	if c.AWSAccessKey == "" || c.AWSSecretAccessKey == "" {
		return aws.Config{}, fmt.Errorf("AWS credentials not found in environment variables")
	}

	creds := credentials.NewStaticCredentialsProvider(
		c.AWSAccessKey,
		c.AWSSecretAccessKey,
		"",
	)

	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(defaultRegion),
		config.WithCredentialsProvider(creds),
	)
	if err != nil {
		return aws.Config{}, fmt.Errorf("failed to load AWS config: %w", err)
	}

	return cfg, nil
}
