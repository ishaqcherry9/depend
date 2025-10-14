package oss

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/ishaqcherry9/depend/pkg/logger"
	jsoniter "github.com/json-iterator/go"
	"golang.org/x/sync/errgroup"
	"io"
	"net/http"
	"os"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	minio "github.com/minio/minio-go/v7"
	minioCredentials "github.com/minio/minio-go/v7/pkg/credentials"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

type AwsClient struct {
	AwsCli    *s3.Client
	AwsBucket string
	AwsAcl    string
}
type MinioClient struct {
	MinioCli    *minio.Client
	MinioBucket string
	MinioAcl    string
}

type Storage struct {
	Aws     AwsClient
	Minio   MinioClient
	Context context.Context
}

type AwsConfig struct {
	AwsAccessKey string
	AwsSecretKey string
	AwsRegion    string
	AwsBucket    string
	AwsAcl       string
}

type MinioConfig struct {
	MinioEndpoint  string
	MinioBucket    string
	MinioAccessKey string
	MinioSecretKey string
	MinioUseSSL    bool
}

type UploadOssInfo struct {
	Key         string `json:"key"`
	FileName    string `json:"filename"`
	Acl         string `json:"acl"`
	ContentType string `json:"content_type"`
}

type Object struct {
	Keys []string `json:"key"`
}

// 理论上应以api提供给业务来解耦，但业务需实现分片上传、断点续传等细节，文件还要2次中转，成本高，以lib方式提供，但配置信息和相关开关对业务透明。
// 通过configAddr获取线下或线上配置信息。
func NewStorage(ctx context.Context, configAddr string) (*Storage, error) {
	awsConf, minioConf, err := getConfig(ctx, configAddr)
	if err != nil {
		logger.Errorf(ctx, "Error getting config: %v", err)
		return nil, err
	}

	/*fmt.Println("aws %+v", awsConf)
	fmt.Println("miniIo %+v", minioConf)*/

	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion(awsConf.AwsRegion),
		config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(awsConf.AwsAccessKey, awsConf.AwsSecretKey, ""),
		),
	)
	if err != nil {
		logger.Errorf(ctx, "无法加载 AWS 配置: %v", err)
		return nil, err
	}

	awsClient := s3.NewFromConfig(cfg)

	minioClient, err := minio.New(minioConf.MinioEndpoint,
		&minio.Options{
			Creds:  minioCredentials.NewStaticV4(minioConf.MinioAccessKey, minioConf.MinioSecretKey, ""),
			Secure: minioConf.MinioUseSSL,
		})
	if err != nil {
		logger.Errorf(ctx, "无法加载 MinIO 配置: %v", err)
		return nil, err
	}
	policy := fmt.Sprintf(`{
		        "Version": "2012-10-17",
		        "Statement": [
		            {
		                "Effect": "Allow",
		                "Principal": {"AWS": ["*"]},
		                "Action": ["s3:GetObject"],
		                "Resource": ["arn:aws:s3:::%s/*"]
		            }
		        ]
		    }`, minioConf.MinioBucket)

	err = minioClient.SetBucketPolicy(ctx, minioConf.MinioBucket, policy)
	if err != nil {
		logger.Errorf(ctx, "设置桶%s 公开失败: %v", minioConf.MinioBucket, err)
	}

	return &Storage{
		Aws: AwsClient{
			AwsCli:    awsClient,
			AwsBucket: awsConf.AwsBucket,
			AwsAcl:    awsConf.AwsAcl,
		},
		Minio: MinioClient{
			MinioCli:    minioClient,
			MinioBucket: minioConf.MinioBucket,
			MinioAcl:    awsConf.AwsAcl,
		},
		Context: ctx,
	}, nil
}

func (s *Storage) Upload(ctx context.Context, info UploadOssInfo) (string, error) {
	var g errgroup.Group
	results := make(map[string]string)
	var mu sync.Mutex

	g.Go(func() error {
		f1, _ := os.Open(info.FileName)
		defer f1.Close()

		uploader := manager.NewUploader(s.Aws.AwsCli)
		out, err := uploader.Upload(ctx, &s3.PutObjectInput{
			Bucket:      aws.String(s.Aws.AwsBucket),
			Key:         aws.String(info.Key),
			Body:        f1,
			ACL:         types.ObjectCannedACL(info.Acl),
			ContentType: aws.String(info.ContentType),
		})

		if err != nil {
			return fmt.Errorf("上传 AWS 失败: %w", err)
		}

		mu.Lock()
		results["aws"] = out.Location
		mu.Unlock()

		return nil
	})

	g.Go(func() error {
		f2, _ := os.Open(info.FileName)
		defer f2.Close()

		out, err := s.Minio.MinioCli.PutObject(ctx, s.Minio.MinioBucket, info.Key, f2, -1, minio.PutObjectOptions{
			ContentType: info.ContentType,
		})
		if err != nil {
			return fmt.Errorf("上传 MinIO 失败: %w", err)
		}

		mu.Lock()
		results["minio"] = out.Location
		mu.Unlock()

		return nil
	})

	if err := g.Wait(); err != nil {
		logger.Errorf(ctx, "UploadConcurrency Wait err: %v", err)
		return "", err
	}

	logger.Infof(ctx, "aws and minio addrs: %v", results)
	return results["aws"], nil
}

func (s *Storage) UploadSeq(ctx context.Context, info UploadOssInfo) (string, error) {
	file, err := os.Open(info.FileName)
	if err != nil {
		logger.Errorf(ctx, "打开文件失败: %v", err)
		return "", err
	}

	defer file.Close()

	uploader := manager.NewUploader(s.Aws.AwsCli)
	awsResult, err := uploader.Upload(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.Aws.AwsBucket),
		Key:         aws.String(info.Key),
		Body:        file,
		ACL:         types.ObjectCannedACL(info.Acl),
		ContentType: aws.String(info.ContentType),
	})
	if err != nil {
		logger.Errorf(ctx, "上传aws失败: %v", err)
		return "", err
	}

	minioResult, err := s.Minio.MinioCli.PutObject(ctx, s.Minio.MinioBucket, info.Key, file, -1, minio.PutObjectOptions{
		ContentType: info.ContentType,
	})
	if err != nil {
		logger.Errorf(ctx, "上传minio失败: %v", err)
		return "", err
	}

	logger.Infof(ctx, "minio location: %s", minioResult.Location)
	return awsResult.Location, nil
}

func (s *Storage) List(ctx context.Context) (Object, error) {
	awsObjects, err := s.Aws.AwsCli.ListObjectsV2(ctx, &s3.ListObjectsV2Input{Bucket: aws.String(s.Aws.AwsBucket)})
	if err != nil {
		logger.Errorf(ctx, "列出 AWS List 失败: %v", err)
		return Object{}, err
	}

	var object Object
	for _, obj := range awsObjects.Contents {
		object.Keys = append(object.Keys, *obj.Key)
	}

	return object, nil

	//s.Minio.MinioCli.ListObjects(ctx, s.Minio.MinioBucket, minio.ListObjectsOptions{Recursive: true})
}

func (s *Storage) Get(ctx context.Context, key string) ([]byte, error) {
	awsObj, err := s.Aws.AwsCli.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.Aws.AwsBucket),
		Key:    aws.String(key),
	})
	if err != nil {
		logger.Errorf(ctx, "从 AWS 获取失败: %v", err)
		return nil, err
	}

	defer awsObj.Body.Close()
	awsData, _ := io.ReadAll(awsObj.Body)
	return awsData, nil

	/*
		minioObj, err := s.Minio.MinioCli.GetObject(ctx, s.Minio.MinioBucket, key, minio.GetObjectOptions{})
		if err != nil {
			return fmt.Errorf("从 MinIO 获取失败: %w", err)
		}

		defer minioObj.Close()
		minioData, _ := io.ReadAll(minioObj)
		fmt.Printf("📦 [MinIO] 内容: %s\n", string(minioData))
	*/
}

func (s *Storage) Delete(ctx context.Context, key string) error {
	_, err := s.Aws.AwsCli.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.Aws.AwsBucket),
		Key:    aws.String(key),
	})
	if err != nil {
		logger.Errorf(ctx, "删除 AWS 失败: %v", err)
		return err
	}

	err = s.Minio.MinioCli.RemoveObject(ctx, s.Minio.MinioBucket, key, minio.RemoveObjectOptions{})
	if err != nil {
		logger.Errorf(ctx, "删除 MinIO 失败: %v", err)
		return err
	}

	return nil
}

func getConfig(ctx context.Context, address string) (AwsConfig, MinioConfig, error) {
	resp, err := http.Get(address)
	if err != nil {
		logger.Errorf(ctx, "oss getConfig 请求错误 %v", err)
		return AwsConfig{}, MinioConfig{}, err
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Errorf(ctx, "oss getConfig 读取响应错误: %v", err)
		return AwsConfig{}, MinioConfig{}, err
	}

	type Result struct {
		AwsConf   AwsConfig   `json:"AwsConfig"`
		MinioConf MinioConfig `json:"MinioConfig"`
	}

	type Response struct {
		Code      int    `json:"code"`
		Msg       string `json:"msg"`
		Data      Result `json:"data"`
		RequestId string `json:"requestId"`
	}

	var result Response
	if err := json.Unmarshal(body, &result); err != nil {
		logger.Errorf(ctx, "oss getConfig 解析json错误: %v", err)
		return AwsConfig{}, MinioConfig{}, err
	}

	return result.Data.AwsConf, result.Data.MinioConf, nil
}
