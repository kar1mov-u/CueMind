package storage

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

type Storage struct {
	s3Client        *s3.Client
	bucketName      string
	s3PresignClient *s3.PresignClient
}

func New(bucketName string) *Storage {
	s3Client, err := newS3Client()
	if err != nil {
		log.Fatalf("Failure on Creating Storage: %v", err)
	}

	presignedClient := s3.NewPresignClient(s3Client)
	return &Storage{s3Client: s3Client, bucketName: bucketName, s3PresignClient: presignedClient}
}

func newS3Client() (*s3.Client, error) {
	ctx := context.Background()
	sdkConfig, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("Cannot load config for AWS_SDK: %v", err)
	}
	s3Client := s3.NewFromConfig(sdkConfig)
	return s3Client, nil
}

func (s *Storage) UploadFile(ctx context.Context, key string, body io.Reader) error {
	_, err := s.s3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(key),
		Body:   body,
	})
	if err != nil {
		return fmt.Errorf("Couldn't upload file %v to %v:%v. Here's why: %v", key, s.bucketName, key, err)
	}

	err = s3.NewObjectExistsWaiter(s.s3Client).Wait(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(key),
	}, time.Minute)

	if err != nil {
		return fmt.Errorf("object exists waiter failed for %s: %w", key, err)
	}
	return nil
}

func (s *Storage) GetFile(ctx context.Context, key string) (io.ReadCloser, error) {
	result, err := s.s3Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(key),
	})
	if err != nil {
		var noKey *types.NoSuchKey
		if errors.As(err, &noKey) {
			log.Printf("Can't get Object %s from bucket %s . No such key exists.\n", key, s.bucketName)
			err = noKey
		} else {
			log.Printf("Couldn't get object %v:%v. Here's why: %v\n", s.bucketName, key, err)
		}
		return nil, err
	}

	return result.Body, nil
}

func (s *Storage) GeneratePresignedUrl(ctx context.Context, key string, lifetime int) (string, error) {
	res, err := s.s3PresignClient.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(key),
	}, func(options *s3.PresignOptions) {
		options.Expires = time.Duration(lifetime * int(time.Minute))
	})
	if err != nil {
		return "", fmt.Errorf("Error on creating PresignedURL: %v", err)
	}
	return res.URL, nil
}

// func (s *Storage) ListFiles()

// func (s *Storage) DeleteFIle()
