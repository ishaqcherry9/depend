package oss

import (
	"bytes"
	"context"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/ishaqcherry9/depend/pkg/logger"
	"io"
	"net/url"
	"sort"
	"sync"
)

type AwsConf struct {
	AwsEndpoint  string `yaml:"awsEndpoint" json:"awsEndpoint"`
	AwsAccessKey string `yaml:"awsAccessKey" json:"awsAccessKey"`
	AwsSecretKey string `yaml:"awsSecretKey" json:"awsSecretKey"`
	AwsRegion    string `yaml:"awsRegion" json:"awsRegion"`
	AwsBucket    string `yaml:"awsBucket" json:"awsBucket"`
	AwsAcl       string `yaml:"awsAcl" json:"awsAcl"`
}

func NewAwsClient(ctx context.Context, conf OssConf) (*s3.Client, error) {
	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion(conf.Aws.AwsRegion),
		config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(conf.Aws.AwsAccessKey, conf.Aws.AwsSecretKey, ""),
		),
	)

	if err != nil {
		logger.Errorf(ctx, "aws config load err: %v", err)
		return nil, err
	}

	return s3.NewFromConfig(cfg), nil
}

// 分片上传，解决当前问题（大文件上传限制和速度慢）
func (s *Storage) AwsUpload(ctx context.Context, reader io.Reader, uploadKey, fileType string) (string, error) {
	createResp, err := s.Aws.CreateMultipartUpload(ctx, &s3.CreateMultipartUploadInput{
		Bucket:      aws.String(s.AwsBucket),
		Key:         aws.String(uploadKey),
		ACL:         s3types.ObjectCannedACLPublicRead,
		ContentType: aws.String(fileType),
	})
	if err != nil {
		logger.Errorf(ctx, "aws Multipart upload err %v", err)
		return "", err
	}

	uploadID := *createResp.UploadId
	logger.Infof(ctx, "aws upload id: %s", uploadID)

	var completedParts []s3types.CompletedPart
	var mu sync.Mutex
	var wg sync.WaitGroup
	partNumber := 1

	for {
		partBuffer := make([]byte, partSize)
		n, err := reader.Read(partBuffer)
		if err != nil && err != io.EOF {
			logger.Errorf(ctx, "aws reader err: %v", err)
			return "", err
		}
		if n == 0 {
			break
		}

		wg.Add(1)
		go func(partNum int, data []byte) {
			defer wg.Done()

			resp, err := s.Aws.UploadPart(ctx, &s3.UploadPartInput{
				Bucket:     aws.String(s.AwsBucket),
				Key:        aws.String(uploadKey),
				PartNumber: aws.Int32(int32(partNum)),
				UploadId:   aws.String(uploadID),
				Body:       bytes.NewReader(data),
			})
			if err != nil {
				logger.Errorf(ctx, "aws uploadID %s part err %v", uploadID, err)
				return
			}

			mu.Lock()
			completedParts = append(completedParts, s3types.CompletedPart{
				ETag:       resp.ETag,
				PartNumber: aws.Int32(int32(partNum)),
			})
			mu.Unlock()
		}(partNumber, partBuffer[:n])

		partNumber++
	}

	wg.Wait()

	sort.Slice(completedParts, func(i, j int) bool {
		return *completedParts[i].PartNumber < *completedParts[j].PartNumber
	})

	completeResp, err := s.Aws.CompleteMultipartUpload(ctx, &s3.CompleteMultipartUploadInput{
		Bucket:   aws.String(s.AwsBucket),
		Key:      aws.String(uploadKey),
		UploadId: aws.String(uploadID),
		MultipartUpload: &s3types.CompletedMultipartUpload{
			Parts: completedParts,
		},
	})
	if err != nil {
		logger.Errorf(ctx, "aws CompleteMultipartUpload err: %v", err)
		return "", err
	}

	*completeResp.Location, _ = url.QueryUnescape(*completeResp.Location)
	return *completeResp.Location, nil
}

func (s *Storage) AwsDelete(ctx context.Context, key string) error {
	_, err := s.Aws.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.AwsBucket),
		Key:    aws.String(key),
	})

	logger.Infof(ctx, "delete aws key %s err msg:%v", s.AwsBucket+"/"+key, err)
	return err
}
