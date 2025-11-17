package oss

import (
	"context"
	"fmt"
	"github.com/ishaqcherry9/depend/pkg/logger"
	minio "github.com/minio/minio-go/v7"
	minioCredentials "github.com/minio/minio-go/v7/pkg/credentials"
	"io"
)

type MinioConf struct {
	MinioAccessKey string `yaml:"minioAccessKey" json:"minioAccessKey"`
	MinioSecretKey string `yaml:"minioSecretKey" json:"minioSecretKey"`
	MinioEndpoint  string `yaml:"minioEndpoint" json:"minioEndpoint"`
	MinioBucket    string `yaml:"minioBucket" json:"minioBucket"`
	MinioUseSSL    bool   `yaml:"minioUseSSL" json:"minioUseSSL"`
}

func NewMinioClient(ctx context.Context, conf OssConf) (*minio.Client, error) {
	minioClient, err := minio.New(conf.Minio.MinioEndpoint,
		&minio.Options{
			Creds:  minioCredentials.NewStaticV4(conf.Minio.MinioAccessKey, conf.Minio.MinioSecretKey, ""),
			Secure: conf.Minio.MinioUseSSL,
		})
	if err != nil {
		logger.Errorf(ctx, "minio config load err: %v", err)
		return nil, err
	}

	//对指定bucket设置action操作（Allow非Deny，默认可Put）。
	//当前为备用通道，权限设置宽，如后续切为线上需严格下。
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
		    }`, conf.Minio.MinioBucket)

	err = minioClient.SetBucketPolicy(ctx, conf.Minio.MinioBucket, policy)
	if err != nil {
		logger.Errorf(ctx, "set bucket %s open err: %v", conf.Minio.MinioBucket, err)
		return nil, err
	}

	return minioClient, nil
}

func (s *Storage) MinioUpload(ctx context.Context, reader io.Reader, uploadKey, fileType string, size int64) (string, error) {
	var num uint
	if size > 0 {
		num = uint(size/partSize) + 1
		if num > 20 {
			num = 20
		}
	} else {
		num = 1
	}

	opts := minio.PutObjectOptions{
		ContentType:           fileType,
		NumThreads:            num,
		PartSize:              partSize,
		ConcurrentStreamParts: true,
	}

	info, err := s.Minio.PutObject(ctx, s.MinioBucket, uploadKey, reader, -1, opts)
	if err != nil {
		logger.Errorf(ctx, "minio upload err: %v", err)
		return "", err
	}

	logger.Infof(ctx, "minio upload success, info.Location: %s", info.Location)
	return info.Location, nil
}

func (s *Storage) MinioDelete(ctx context.Context, key string) error {
	err := s.Minio.RemoveObject(ctx, s.MinioBucket, key, minio.RemoveObjectOptions{})
	logger.Infof(ctx, "delete minio key %s err msg: %v", s.MinioBucket+"/"+key, err)
	return err
}
