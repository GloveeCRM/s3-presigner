package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"s3-presigner/internal/config"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type Storage struct {
	client        *s3.Client
	presignClient *s3.PresignClient
	awsConfig     aws.Config
	gloveeAPIKey  string
}

func New(cfg *config.Config) (*Storage, error) {
	awsConfig, err := cfg.LoadAWSConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	client := s3.NewFromConfig(awsConfig)
	presignClient := s3.NewPresignClient(client)

	return &Storage{
		client:        client,
		presignClient: presignClient,
		awsConfig:     awsConfig,
		gloveeAPIKey:  cfg.GloveeAPIKey,
	}, nil
}

func ValidateAWSCredentials(cfg *config.Config) error {
	storage, err := New(cfg)
	if err != nil {
		return fmt.Errorf("failed to initialize AWS client: %w", err)
	}

	_, err = storage.client.ListBuckets(context.TODO(), &s3.ListBucketsInput{})
	if err != nil {
		return fmt.Errorf("invalid AWS credentials: %w", err)
	}

	return nil
}

func (s *Storage) GetObjectPresignedURL(region, bucket, objectKey string, expiresIn int64) (string, error) {
	cfg := s.getRegionConfig(region)
	client := s3.NewFromConfig(cfg)
	presignClient := s3.NewPresignClient(client)

	request, err := presignClient.PresignGetObject(context.TODO(),
		&s3.GetObjectInput{
			Bucket: &bucket,
			Key:    &objectKey,
		},
		func(po *s3.PresignOptions) {
			po.Expires = time.Second * time.Duration(expiresIn)
		},
	)
	if err != nil {
		return "", fmt.Errorf("error presigning GET object: %w", err)
	}

	return request.URL, nil
}

func (s *Storage) PutObjectPresignedURL(region, bucket, objectKey string, expiresIn int64) (string, error) {
	cfg := s.getRegionConfig(region)
	client := s3.NewFromConfig(cfg)
	presignClient := s3.NewPresignClient(client)

	request, err := presignClient.PresignPutObject(context.TODO(),
		&s3.PutObjectInput{
			Bucket: &bucket,
			Key:    &objectKey,
		},
		func(po *s3.PresignOptions) {
			po.Expires = time.Second * time.Duration(expiresIn)
		},
	)
	if err != nil {
		return "", fmt.Errorf("error presigning PUT object: %w", err)
	}

	return request.URL, nil
}

func (s *Storage) DeleteObjectPresignedURL(region, bucket, objectKey string, expiresIn int64) (string, error) {
	cfg := s.getRegionConfig(region)
	client := s3.NewFromConfig(cfg)
	presignClient := s3.NewPresignClient(client)

	request, err := presignClient.PresignDeleteObject(context.TODO(),
		&s3.DeleteObjectInput{
			Bucket: &bucket,
			Key:    &objectKey,
		},
		func(po *s3.PresignOptions) {
			po.Expires = time.Second * time.Duration(expiresIn)
		},
	)
	if err != nil {
		return "", fmt.Errorf("error presigning DELETE object: %w", err)
	}

	return request.URL, nil
}

func (s *Storage) ObjectExists(region, bucket, objectKey string) error {
	cfg := s.getRegionConfig(region)
	client := s3.NewFromConfig(cfg)
	_, err := client.HeadObject(context.TODO(), &s3.HeadObjectInput{
		Bucket: &bucket,
		Key:    &objectKey,
	})
	if err != nil {
		return fmt.Errorf("object does not exist or is not accessible: %w", err)
	}
	return nil
}

func (s *Storage) getRegionConfig(region string) aws.Config {
	cfg := s.awsConfig
	cfg.Region = region
	return cfg
}

type FileDetails struct {
	FileID    int64  `json:"file_id"`
	Region    string `json:"region"`
	Bucket    string `json:"bucket"`
	ObjectKey string `json:"object_key"`
}

func (s *Storage) GetFileDetails(fileID, userID string) (FileDetails, error) {
	baseURL := "https://api.glovee.io/rpc/file_details"
	params := url.Values{}
	params.Add("file_id", fileID)
	params.Add("user_id", userID)

	requestURL := fmt.Sprintf("%s?%s", baseURL, params.Encode())

	req, err := http.NewRequest(http.MethodGet, requestURL, nil)
	if err != nil {
		return FileDetails{}, fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", s.gloveeAPIKey))

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return FileDetails{}, fmt.Errorf("error sending request: %w", err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return FileDetails{}, fmt.Errorf("error reading file details: %w", err)
	}

	if res.StatusCode != http.StatusOK {
		return FileDetails{}, fmt.Errorf("error getting file details: status code %d", res.StatusCode)
	}

	var fileDetails FileDetails
	err = json.Unmarshal(body, &fileDetails)
	if err != nil {
		return FileDetails{}, fmt.Errorf("error unmarshalling file details: %w", err)
	}

	return fileDetails, nil
}
