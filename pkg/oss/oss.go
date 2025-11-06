package oss

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/disintegration/imaging"
	json "github.com/ishaqcherry9/depend/pkg/json-iterator"
	"github.com/ishaqcherry9/depend/pkg/logger"
	minio "github.com/minio/minio-go/v7"
	minioCredentials "github.com/minio/minio-go/v7/pkg/credentials"
	ffmpeg "github.com/u2takey/ffmpeg-go"
	"golang.org/x/sync/errgroup"
	"io"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// 上传大小限制，后续需分片上传
const MaxUploadSize = 20 * 1024 * 1024

type AwsConf struct {
	AwsAccessKey string `yaml:"awsAccessKey" json:"awsAccessKey"`
	AwsSecretKey string `yaml:"awsSecretKey" json:"awsSecretKey"`
	AwsRegion    string `yaml:"awsRegion" json:"awsRegion"`
	AwsBucket    string `yaml:"awsBucket" json:"awsBucket"`
	AwsAcl       string `yaml:"awsAcl" json:"awsAcl"`
}

type MinioConf struct {
	MinioAccessKey string `yaml:"minioAccessKey" json:"minioAccessKey"`
	MinioSecretKey string `yaml:"minioSecretKey" json:"minioSecretKey"`
	MinioEndpoint  string `yaml:"minioEndpoint" json:"minioEndpoint"`
	MinioBucket    string `yaml:"minioBucket" json:"minioBucket"`
	MinioUseSSL    bool   `yaml:"minioUseSSL" json:"minioUseSSL"`
}

type OssConf struct {
	TmpDir   string    `yaml:"tmpDir" json:"tmpDir"`
	Provider string    `yaml:"provider" json:"provider"`
	Aws      AwsConf   `yaml:"aws" json:"aws"`
	Minio    MinioConf `yaml:"minio" json:"minio"`
}

type Storage struct {
	Aws         *s3.Client    `json:"aws_cli"`
	AwsBucket   string        `json:"aws_bucket"`
	Minio       *minio.Client `json:"minio_cli"`
	MinioBucket string        `json:"minio_bucket"`

	Context context.Context `json:"context"`

	SiteID   string `json:"site_id"`
	TmpDir   string `yaml:"tmpDir" json:"tmpDir"`
	Provider string `yaml:"provider" json:"provider"`
}

func NewStorage(ctx context.Context, ossConf OssConf, siteID string) (error, *Storage) {
	if ossConf.TmpDir == "" || ossConf.Aws.AwsBucket == "" || ossConf.Minio.MinioBucket == "" {
		logger.Errorf(ctx, "get storage config err")
		return errors.New("get storage config err"), nil
	}

	//aws
	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion(ossConf.Aws.AwsRegion),
		config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(ossConf.Aws.AwsAccessKey, ossConf.Aws.AwsSecretKey, ""),
		),
	)
	if err != nil {
		logger.Errorf(ctx, "aws config load err: %v", err)
		return err, nil
	}

	awsClient := s3.NewFromConfig(cfg)

	//minio
	minioClient, err := minio.New(ossConf.Minio.MinioEndpoint,
		&minio.Options{
			Creds:  minioCredentials.NewStaticV4(ossConf.Minio.MinioAccessKey, ossConf.Minio.MinioSecretKey, ""),
			Secure: ossConf.Minio.MinioUseSSL,
		})
	if err != nil {
		logger.Errorf(ctx, "minio config load err: %v", err)
		return err, nil
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
		    }`, ossConf.Minio.MinioBucket)

	err = minioClient.SetBucketPolicy(ctx, ossConf.Minio.MinioBucket, policy)
	if err != nil {
		logger.Errorf(ctx, "set bucket %s open err: %v", ossConf.Minio.MinioBucket, err)
		return err, nil
	}

	stroage := &Storage{
		Aws:         awsClient,
		AwsBucket:   ossConf.Aws.AwsBucket,
		Minio:       minioClient,
		MinioBucket: ossConf.Minio.MinioBucket,
		TmpDir:      ossConf.TmpDir,
		Provider:    ossConf.Provider,
		SiteID:      siteID,
		Context:     ctx,
	}

	logger.Infof(ctx, "set storage config success: %v", stroage)
	return nil, stroage
}

type UploadRequest struct {
	//必须
	BussType   string `json:"buss_type"`
	FileName   string `json:"file_name"`
	UploadType int8   `json:"upload_type"`

	//非必须
	Acl           string `json:"acl"`
	NeedCover     bool   `json:"need_cover"`
	ThumbWidth    int    `json:"thumb_width"`
	ThumbHeight   int    `json:"thumb_height"`
	NeedWaterMask bool   `json:"need_water_mask"`
	FileType      string `json:"file_type"`
	UniqID        string `json:"uniq_id"`
}

type UploadResponse struct {
	Location   string `json:"location"`
	Thumbnail  string `json:"thumbnail"`
	Height     int    `json:"height"`
	Width      int    `json:"width"`
	VideoCover string `json:"video_cover"`
	Duration   int64  `json:"duration"`
}

// 文件直接上传到aws和minio，不涉及图片视频再加工。
func (s *Storage) UploadDirect(ctx context.Context, info UploadRequest, r *http.Request) (UploadResponse, error) {
	file, handler, err := r.FormFile("file")
	if err != nil {
		logger.Errorf(ctx, "get file err: %v", err)
		return UploadResponse{}, err
	}
	defer file.Close()

	if handler.Size > MaxUploadSize {
		return UploadResponse{}, errors.New("file too big")
	}

	date := time.Now().Format("20060102")
	uploadKey := filepath.Join(s.SiteID, info.BussType, date, uniqueID())
	if info.Acl == "" {
		info.Acl = "public-read"
	}

	//并发读取，防止读到空流
	buf, _ := io.ReadAll(file)
	reader1 := io.NopCloser(bytes.NewReader(buf))
	reader2 := io.NopCloser(bytes.NewReader(buf))

	result := UploadResponse{}
	var g errgroup.Group

	g.Go(func() error {
		uploader := manager.NewUploader(s.Aws)
		out, err := uploader.Upload(ctx, &s3.PutObjectInput{
			Bucket:      aws.String(s.AwsBucket),
			Key:         aws.String(uploadKey),
			Body:        reader1,
			ACL:         types.ObjectCannedACL(info.Acl),
			ContentType: aws.String(handler.Header.Get("Content-Type")),
		})
		if err != nil {
			logger.Errorf(ctx, "aws upload err: %v", err)
			return err
		}

		result.Location = out.Location
		return nil
	})

	g.Go(func() error {
		outMinio, err := s.Minio.PutObject(ctx, s.MinioBucket, uploadKey, reader2, -1, minio.PutObjectOptions{
			ContentType: info.FileType,
		})
		if err != nil {
			logger.Errorf(ctx, "MinIO upload err: %v", err)
			//暂时也返err，防止后续切换，数据不同步。
			return err
		}

		logger.Infof(ctx, "minio upload addr: ", outMinio.Location)
		return nil
	})

	if err := g.Wait(); err != nil {
		logger.Errorf(ctx, "UploadConcurrency Wait err: %v", err)
		return UploadResponse{}, err
	}

	return result, nil
}

func (s *Storage) ImageWork(ctx context.Context, info UploadRequest, result *UploadResponse) error {
	img, err := imaging.Open(info.FileName)
	if err != nil {
		logger.Errorf(s.Context, "open file err: ", err)
		return err
	}

	x := img.Bounds()
	result.Width = x.Max.X
	result.Height = x.Max.Y

	if info.ThumbWidth > 0 && info.ThumbHeight > 0 {
		thumbFileName := info.FileName + "-thumb.jpg"
		thumb := imaging.Resize(img, info.ThumbWidth, info.ThumbHeight, imaging.Lanczos)
		options := imaging.JPEGQuality(80)

		err = imaging.Save(thumb, thumbFileName, options)
		if err != nil {
			logger.Errorf(ctx, "save thumb err: %v", err)
			return err
		}

		result.Thumbnail = thumbFileName
		logger.Infof(ctx, "save thumb addr: ", result.Thumbnail)
	}

	return nil
}

func (s *Storage) VideoWork(ctx context.Context, info UploadRequest, result *UploadResponse) error {
	result.Duration = s.GetDuration(info)

	if info.NeedCover {
		coverFileName := info.FileName + "-cover.jpg"
		timeSec := 5

		err := ffmpeg.Input(info.FileName, ffmpeg.KwArgs{"ss": timeSec}).Output(coverFileName, ffmpeg.KwArgs{
			"vframes": 1,
			"q:v":     2,
		}).Run()

		if err != nil {
			logger.Errorf(s.Context, "ffmpeg.Input err %v", err)
			return err
		}

		result.VideoCover = coverFileName
	}

	return nil
}

func (s *Storage) AudioWork(ctx context.Context, info UploadRequest, result *UploadResponse) error {
	result.Duration = s.GetDuration(info)
	return nil
}

func (s *Storage) DeleteTmpFile(ctx context.Context, fileNames []string) error {
	var errs []error

	for _, name := range fileNames {
		if name != "" {
			if err := os.Remove(name); err != nil {
				logger.Errorf(ctx, "Delete file %s err %v", name, err)
				errs = append(errs, err)
			}
		}
	}

	return errors.Join(errs...)
}

// 除上传功能外，图片会返回长宽、视频音频会返回时长这些额外信息。
// 根据图片thumb设置返回缩略图，根据视频cover设置返回某些视频帧合集。
// 先本地存储，获取长宽|时长额外信息，裁剪图片视频后，并发上传到aws和MinIO。
func (s *Storage) UploadExtra(ctx context.Context, info UploadRequest, r *http.Request) (UploadResponse, error) {
	file, handler, err := r.FormFile("file")
	if err != nil {
		logger.Errorf(ctx, "get file err: %v", err)
		return UploadResponse{}, err
	}
	defer file.Close()

	if handler.Size > MaxUploadSize {
		return UploadResponse{}, errors.New("file too big")
	}

	info.UniqID = uniqueID()
	os.MkdirAll(s.TmpDir, os.ModePerm)
	savePath := filepath.Join(s.TmpDir, info.UniqID)
	dst, err := os.Create(savePath)
	if err != nil {
		logger.Errorf(ctx, "create file err: %v", err)
		return UploadResponse{}, err
	}

	defer dst.Close()
	if _, err := io.Copy(dst, file); err != nil {
		return UploadResponse{}, err
	}

	info.FileName = savePath
	result := UploadResponse{}
	switch info.UploadType {
	case 1:
		err = s.ImageWork(ctx, info, &result)
	case 2:
		err = s.VideoWork(ctx, info, &result)
	case 3:
		err = s.AudioWork(ctx, info, &result)
	default:
		logger.Errorf(ctx, "unknown upload type: %v", info.UploadType)
	}

	fileNames := []string{info.FileName, result.Thumbnail, result.VideoCover}
	defer s.DeleteTmpFile(ctx, fileNames)

	if err != nil {
		logger.Errorf(ctx, "upload err: %v", err)
		return UploadResponse{}, err
	}

	if info.Acl == "" {
		info.Acl = "public-read"
	}

	var g errgroup.Group
	g.Go(func() error {
		err := s.AwsUpload(ctx, info, &result)
		if err != nil {
			logger.Errorf(ctx, "aws upload err: %v", err)
			return err
		}
		return nil
	})

	g.Go(func() error {
		err := s.MinioUpload(ctx, info, result)
		if err != nil {
			logger.Errorf(ctx, "aws upload err: %v", err)
			return err
		}
		return nil
	})

	if err := g.Wait(); err != nil {
		logger.Errorf(ctx, "UploadConcurrency Wait err: %v", err)
		return UploadResponse{}, err
	}

	logger.Infof(ctx, "aws and minio addrs: %+v", result)
	return result, nil
}

func (s *Storage) MinioUpload(ctx context.Context, info UploadRequest, result UploadResponse) error {
	date := time.Now().Format("20060102")
	uploadKey := filepath.Join(s.SiteID, info.BussType, date, info.UniqID)
	location, err := s.minioUpload(ctx, uploadKey, info.FileName, info.FileType)
	if err != nil {
		logger.Errorf(ctx, "MinioUpload fileName %s err: %v", info.FileName, err)
		return err
	}
	logger.Infof(ctx, "minio upload addr: %s", location)

	if result.Thumbnail != "" {
		uploadKey = filepath.Join(s.SiteID, info.BussType, date, info.UniqID+"-thumb")
		location, err := s.minioUpload(ctx, uploadKey, result.Thumbnail, info.FileType)
		if err != nil {
			logger.Errorf(ctx, "MinioUpload Thumbnail err: %v", err)
			return err
		}
		logger.Infof(ctx, "minio upload Thumbnail addr: %s", location)
	}

	if result.VideoCover != "" {
		uploadKey = filepath.Join(s.SiteID, info.BussType, date, info.UniqID+"-cover")
		location, err := s.minioUpload(ctx, uploadKey, result.VideoCover, info.FileType)
		if err != nil {
			logger.Errorf(ctx, "MinioUpload Thumbnail err: %v", err)
			return err
		}
		logger.Infof(ctx, "minio upload VideoCover addr: %s", location)
	}

	return nil
}

func (s *Storage) AwsUpload(ctx context.Context, info UploadRequest, result *UploadResponse) error {
	date := time.Now().Format("20060102")
	uploadKey := filepath.Join(s.SiteID, info.BussType, date, info.UniqID)
	location, err := s.awsUpload(ctx, uploadKey, info.FileName, info.FileType, info.Acl)
	if err != nil {
		logger.Errorf(ctx, "AwsUpload fileName %s err: %v", info.FileName, err)
		return err
	}
	result.Location = location

	if result.Thumbnail != "" {
		uploadKey = filepath.Join(s.SiteID, info.BussType, date, info.UniqID+"-thumb")
		location, err := s.awsUpload(ctx, uploadKey, result.Thumbnail, info.FileType, info.Acl)
		if err != nil {
			logger.Errorf(ctx, "AwsUpload Thumbnail err: %v", err)
			return err
		}
		result.Thumbnail = location
	}

	if result.VideoCover != "" {
		uploadKey = filepath.Join(s.SiteID, info.BussType, date, info.UniqID+"-cover")
		location, err := s.awsUpload(ctx, uploadKey, result.VideoCover, info.FileType, info.Acl)
		if err != nil {
			logger.Errorf(ctx, "AwsUpload Thumbnail err: %v", err)
			return err
		}
		result.VideoCover = location
	}

	return nil
}

func (s *Storage) awsUpload(ctx context.Context, uploadKey, fileName, fileType, acl string) (string, error) {
	f1, _ := os.Open(fileName)
	defer f1.Close()

	uploader := manager.NewUploader(s.Aws)
	out, err := uploader.Upload(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.AwsBucket),
		Key:         aws.String(uploadKey),
		Body:        f1,
		ACL:         types.ObjectCannedACL(acl),
		ContentType: aws.String(fileType),
	})

	if err != nil {
		return "", fmt.Errorf("aws upload err: %w", err)
	}

	return out.Location, nil
}

func (s *Storage) minioUpload(ctx context.Context, uploadKey string, fileName, fileType string) (string, error) {
	f2, _ := os.Open(fileName)
	defer f2.Close()

	key := strings.ReplaceAll(uploadKey, s.MinioBucket+"/", "")
	out, err := s.Minio.PutObject(ctx, s.MinioBucket, key, f2, -1, minio.PutObjectOptions{
		ContentType: fileType,
	})
	if err != nil {
		return "", fmt.Errorf("upload minio err : %w", err)
	}

	return out.Location, nil
}

func (s *Storage) GetDuration(info UploadRequest) int64 {
	probeOutput, err := ffmpeg.Probe(info.FileName)
	if err != nil {
		logger.Errorf(s.Context, "GetDuration err: %v", err)
		return 0
	}

	var data map[string]interface{}
	if err := json.Unmarshal([]byte(probeOutput), &data); err != nil {
		logger.Errorf(s.Context, "GetDuration err: %v", err)
		return 0
	}

	format, ok := data["format"].(map[string]interface{})
	if !ok {
		logger.Errorf(s.Context, "data format not ok")
		return 0
	}

	durationStr, ok := format["duration"].(string)
	if !ok {
		logger.Errorf(s.Context, "data format duration not ok")
		return 0
	}

	durationFloat64, err := strconv.ParseFloat(durationStr, 64)
	if err != nil {
		logger.Errorf(s.Context, "GetDuration err: %v", err)
		return 0
	}

	return int64(durationFloat64)
}

func (s *Storage) WaterMask(info UploadRequest) {
	//imaging.Fill会保持宽高比，imaging.Resize不会
	/*
		watermarkPath := s.TmpDir + "water.jpg"
		watermark, err := imaging.Open(watermarkPath)
		if err != nil {
			return err
		}

		// 计算水印位置 (右下角)
		pos := image.Pt(
			img.Bounds().Dx()-watermark.Bounds().Dx()-10,
			img.Bounds().Dy()-watermark.Bounds().Dy()-10,
		)

		// 叠加水印 (0.9透明度)
		outputPath := s.TmpDir + info.OutputName + "result.jpg"
		dst := imaging.Overlay(img, watermark, pos, 0.9)
		err = imaging.Save(dst, outputPath)
		if err != nil {
			return err
		}
	*/
}

// todo 雪花
func uniqueID() string {
	src := rand.NewSource(time.Now().UnixNano())
	r := rand.New(src)

	uniqid := fmt.Sprintf("%x", time.Now().UnixNano())
	randomNum := r.Intn(9000000) + 1000000

	return fmt.Sprintf("%s%d", uniqid, randomNum)
}

func (s *Storage) getFileType(fileName string) string {
	/*
		contentType := header.Header.Get("Content-Type")
		if contentType != "" && contentType != "application/octet-stream" {
			return contentType
		}

		尝试读取文件头前 512 字节来检测 MIME 类型
		buffer := make([]byte, 512)
		n, _ := file.Read(buffer)
		file.Seek(0, 0) // 记得重置指针，否则后续读取会丢失数据

		detected := http.DetectContentType(buffer[:n])
		if detected != "application/octet-stream" {
			return detected
		}

		根据文件扩展名兜底判断
		ext := strings.ToLower(filepath.Ext(header.Filename))
		switch ext {
		case ".jpg", ".jpeg", ".png", ".gif", ".bmp", ".webp":
			return "image/" + strings.TrimPrefix(ext, ".")
		case ".mp4", ".mov", ".avi", ".mkv", ".flv", ".wmv":
			return "video/" + strings.TrimPrefix(ext, ".")
		case ".mp3", ".wav", ".aac", ".ogg":
			return "audio/" + strings.TrimPrefix(ext, ".")
		case ".pdf":
			return "application/pdf"
		case ".zip", ".rar", ".7z":
			return "application/zip"
		default:
			return "application/octet-stream"
		}*/

	return ""
}

func (s *Storage) Delete(ctx context.Context, key string) error {
	_, errAws := s.Aws.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.AwsBucket),
		Key:    aws.String(key),
	})
	logger.Infof(ctx, "delete aws key %s err msg:%v", s.AwsBucket+"/"+key, errAws)

	errMinio := s.Minio.RemoveObject(ctx, s.MinioBucket, key, minio.RemoveObjectOptions{})
	logger.Infof(ctx, "delete minio key %s err msg: %v", s.MinioBucket+"/"+key, errMinio)

	if s.Provider == "aws" {
		return errAws
	}

	return errMinio
}
