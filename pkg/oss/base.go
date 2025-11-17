package oss

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/disintegration/imaging"
	filetype "github.com/h2non/filetype"
	"github.com/ishaqcherry9/depend/pkg/logger"
	json "github.com/json-iterator/go"
	"github.com/minio/minio-go/v7"
	ffmpeg "github.com/u2takey/ffmpeg-go"
	"golang.org/x/sync/errgroup"
	"image"
	"io"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

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
	Size       int64  `json:"size"`
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

// 每个分片至少5M
const partSize = 5 * 1024 * 1024

type OSS interface {
	UploadFile(ctx context.Context, info UploadRequest, r *http.Request) (UploadResponse, error)
	DeleteFile(ctx context.Context, key string) error
}

// 需要双写
func NewStorage(ctx context.Context, ossConf OssConf, siteID string) (*Storage, error) {
	if ossConf.TmpDir == "" || ossConf.Aws.AwsBucket == "" || ossConf.Minio.MinioBucket == "" {
		logger.Errorf(ctx, "get storage config err")
		return nil, errors.New("get storage config err")
	}

	awsClient, err := NewAwsClient(ctx, ossConf)
	if err != nil {
		logger.Errorf(ctx, "get aws client err: %v", err)
		return nil, err
	}

	minioClient, err := NewMinioClient(ctx, ossConf)
	if err != nil {
		logger.Errorf(ctx, "get minio client err: %v", err)
		return nil, err
	}

	stroage := &Storage{
		Aws:         awsClient,
		AwsBucket:   ossConf.Aws.AwsBucket,
		Minio:       minioClient,
		MinioBucket: ossConf.Minio.MinioBucket,
		TmpDir:      ossConf.TmpDir,
		Provider:    ossConf.Provider,
		SiteID:      "site_" + siteID,
		Context:     ctx,
	}

	logger.Infof(ctx, "set storage config success: %v", stroage)
	return stroage, nil
}

// 文件直接上传到aws和minio，不涉及图片视频二次加工
func (s *Storage) UploadDirect(ctx context.Context, r *http.Request, info UploadRequest) (UploadResponse, error) {
	file, handler, err := r.FormFile("file")
	if err != nil {
		logger.Errorf(ctx, "get file err: %v", err)
		return UploadResponse{}, err
	}
	defer file.Close()

	date := time.Now().Format("20060102")
	uploadKey := filepath.Join(s.SiteID, info.BussType, date, uniqueID())
	result := UploadResponse{
		Size: handler.Size,
	}

	var g errgroup.Group

	//并发读取，防止读到空流
	buf, _ := io.ReadAll(file)
	reader1 := io.NopCloser(bytes.NewReader(buf))
	reader2 := io.NopCloser(bytes.NewReader(buf))

	info.FileType = s.GetFileType(buf)
	logger.Infof(ctx, " from filetype.Match ", info.FileType)
	s.CheckFileType(ctx, info)

	g.Go(func() error {
		result.Location, err = s.AwsUpload(ctx, reader1, uploadKey, info.FileType)
		return err
	})

	g.Go(func() error {
		_, err := s.MinioUpload(ctx, reader2, uploadKey, info.FileType, handler.Size)
		return err
	})

	if err := g.Wait(); err != nil {
		logger.Errorf(ctx, "UploadConcurrency Wait err: %v", err)
		return UploadResponse{}, err
	}

	return result, nil
}

// todo 雪花
func uniqueID() string {
	src := rand.NewSource(time.Now().UnixNano())
	r := rand.New(src)

	uniqid := fmt.Sprintf("%x", time.Now().UnixNano())
	randomNum := r.Intn(9000000) + 1000000

	return fmt.Sprintf("%s%d", uniqid, randomNum)
}

func (s *Storage) ImageExecute(ctx context.Context, info UploadRequest, result *UploadResponse) error {
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

func (s *Storage) VideoExecute(ctx context.Context, info UploadRequest, result *UploadResponse) error {
	result.Duration = s.GetDuration(info)

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

	return nil
}

func (s *Storage) AudioExecute(ctx context.Context, info UploadRequest, result *UploadResponse) error {
	result.Duration = s.GetDuration(info)
	return nil
}

// 临时存储客户端上传文件，用于二次加工
func (s *Storage) Prepare(ctx context.Context, r *http.Request, info *UploadRequest, result *UploadResponse) error {
	file, handle, err := r.FormFile("file")
	if err != nil {
		logger.Errorf(ctx, "get file err: %v", err)
		return err
	}

	defer file.Close()

	info.UniqID = uniqueID()
	os.MkdirAll(s.TmpDir, os.ModePerm)
	savePath := filepath.Join(s.TmpDir, info.UniqID)
	dst, err := os.Create(savePath)
	if err != nil {
		logger.Errorf(ctx, "create file err: %v", err)
		return err
	}

	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		logger.Errorf(ctx, "copy file err: %v", err)
		return err
	}

	dst2, err := os.Open(savePath)
	defer dst2.Close()

	bufHeader := make([]byte, 512)
	n, _ := dst2.Read(bufHeader)
	uploadFileType := s.GetFileType(bufHeader[:n])

	info.FileType = uploadFileType
	info.FileName = savePath
	result.Size = handle.Size
	return nil
}

func (s *Storage) Execute(ctx context.Context, info UploadRequest, result *UploadResponse) error {
	var err error
	switch info.UploadType {
	case 1:
		err = s.ImageExecute(ctx, info, result)
	case 2:
		err = s.VideoExecute(ctx, info, result)
	case 3:
		err = s.AudioExecute(ctx, info, result)
	default:
		logger.Errorf(ctx, "unknown upload type: %v", info.UploadType)
	}

	if err != nil {
		logger.Errorf(ctx, "info.UploadType %d Execute err: %v", info.UploadType, err)
		return err
	}

	return nil
}

func (s *Storage) AwsUploadALL(ctx context.Context, info UploadRequest, result *UploadResponse) error {
	var err error
	date := time.Now().Format("20060102")
	uploadKey := filepath.Join(s.SiteID, info.BussType, date, info.UniqID)

	result.Location, err = s.AwsUploadOne(ctx, info.FileName, uploadKey, info.FileType)
	if err != nil {
		logger.Errorf(ctx, "aws upload info.FileName %s  err: %v", info.FileName, err)
		return err
	}

	if result.Thumbnail != "" {
		uploadKey = filepath.Join(s.SiteID, info.BussType, date, info.UniqID+"-thumb")
		result.Thumbnail, err = s.AwsUploadOne(ctx, result.Thumbnail, uploadKey, info.FileType)
		if err != nil {
			logger.Errorf(ctx, "aws upload Thumbnail %s  err: %v", result.Thumbnail, err)
			return err
		}
	}

	if result.VideoCover != "" {
		uploadKey = filepath.Join(s.SiteID, info.BussType, date, info.UniqID+"-cover")
		result.VideoCover, err = s.AwsUploadOne(ctx, result.VideoCover, uploadKey, info.FileType)
		if err != nil {
			logger.Errorf(ctx, "aws upload VideoCover %s  err: %v", result.VideoCover, err)
			return err
		}
	}

	return nil
}

func (s *Storage) AwsUploadOne(ctx context.Context, fileName, uploadKey, fileType string) (string, error) {
	file, _ := os.Open(fileName)
	defer file.Close()

	buf, _ := io.ReadAll(file)
	reader := io.NopCloser(bytes.NewReader(buf))

	location, err := s.AwsUpload(ctx, reader, uploadKey, fileType)
	if err != nil {
		return "", err
	}

	return location, nil
}

func (s *Storage) MinioUploadOne(ctx context.Context, fileName, uploadKey, fileType string, size int64) (string, error) {
	file, _ := os.Open(fileName)
	defer file.Close()

	buf, _ := io.ReadAll(file)
	reader := io.NopCloser(bytes.NewReader(buf))

	location, err := s.MinioUpload(ctx, reader, uploadKey, fileType, size)
	if err != nil {
		return "", err
	}

	return location, nil
}

func (s *Storage) Finish(ctx context.Context, info UploadRequest, result *UploadResponse) error {
	var g errgroup.Group

	g.Go(func() error {
		if err := s.AwsUploadALL(ctx, info, result); err != nil {
			logger.Errorf(ctx, "aws AwsUploadALL err: %v", err)
			return err
		}
		return nil
	})

	g.Go(func() error {
		if err := s.MinioUploadALL(ctx, info, result); err != nil {
			logger.Errorf(ctx, "minio MinioUploadALL err: %v", err)
			return err
		}
		return nil
	})

	if err := g.Wait(); err != nil {
		logger.Errorf(ctx, "UploadConcurrency Wait err: %v", err)
		return err
	}

	logger.Infof(ctx, "aws and minio addrs: %+v", result)
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

// 加工图片、视频、音频：
// 图片缩率图、原始尺寸、图片大小。
// 视频封面图、视频时长、视频大小。
// 音频时长、音频大小。
// Prepare-Execute-Finish
func (s *Storage) UploadExtra(ctx context.Context, r *http.Request, info UploadRequest) (UploadResponse, error) {
	result := UploadResponse{}

	if err := s.Prepare(ctx, r, &info, &result); err != nil {
		logger.Errorf(ctx, "prepare err: %v", err)
		return result, err
	}

	fileNames := []string{info.FileName, result.Thumbnail, result.VideoCover}
	defer s.DeleteTmpFile(ctx, fileNames)

	s.CheckFileType(ctx, info)

	if err := s.Execute(ctx, info, &result); err != nil {
		logger.Errorf(ctx, "Execute err: %v", err)
		return result, err
	}

	if err := s.Finish(ctx, info, &result); err != nil {
		logger.Errorf(ctx, "Finish err: %v", err)
		return result, err
	}

	return result, nil
}

// 类型判断
func (s *Storage) CheckFileType(ctx context.Context, info UploadRequest) bool {
	logger.Infof(ctx, "Check file type UploadTypeID %d Detect %s", info.UploadType, info.FileType)

	if info.UploadType == 1 && !strings.HasPrefix(info.FileType, "image") {
		logger.Errorf(ctx, "upload type: %v, info.FileType %v", info.UploadType, info.FileType)
		return false
	}
	if info.UploadType == 2 && !strings.HasPrefix(info.FileType, "video") {
		logger.Errorf(ctx, "upload type: %v, info.FileType %v", info.UploadType, info.FileType)
		return false
	}
	if info.UploadType == 3 && !strings.HasPrefix(info.FileType, "audio") {
		logger.Errorf(ctx, "upload type: %v, info.FileType %v", info.UploadType, info.FileType)
		return false
	}

	return true
}

func (s *Storage) MinioUploadALL(ctx context.Context, info UploadRequest, result *UploadResponse) error {
	date := time.Now().Format("20060102")
	uploadKey := filepath.Join(s.SiteID, info.BussType, date, info.UniqID)
	var err error
	_, err = s.MinioUploadOne(ctx, info.FileName, uploadKey, info.FileType, result.Size)
	if err != nil {
		logger.Errorf(ctx, "MinioUpload fileName %s err: %v", info.FileName, err)
		return err
	}

	if result.Thumbnail != "" {
		uploadKey = filepath.Join(s.SiteID, info.BussType, date, info.UniqID+"-thumb")
		_, err = s.MinioUploadOne(ctx, result.Thumbnail, uploadKey, info.FileType, 0)
		if err != nil {
			logger.Errorf(ctx, "MinioUpload Thumbnail err: %v", err)
			return err
		}
	}

	if result.VideoCover != "" {
		uploadKey = filepath.Join(s.SiteID, info.BussType, date, info.UniqID+"-cover")
		_, err = s.MinioUploadOne(ctx, result.VideoCover, uploadKey, info.FileType, 0)
		if err != nil {
			logger.Errorf(ctx, "MinioUpload Thumbnail err: %v", err)
			return err
		}
	}

	return nil
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

func (s *Storage) WaterMask(info UploadRequest) error {
	//imaging.Fill会保持宽高比，imaging.Resize不会
	watermarkPath := s.TmpDir + "water.jpg"
	watermark, err := imaging.Open(watermarkPath)
	if err != nil {
		logger.Errorf(s.Context, "WaterMask err: %v", err)
		return err
	}
	img, err := imaging.Open(info.FileName)
	if err != nil {
		logger.Errorf(s.Context, "WaterMask err: %v", err)
		return err
	}

	// 计算水印位置 (右下角)
	pos := image.Pt(
		img.Bounds().Dx()-watermark.Bounds().Dx()-10,
		img.Bounds().Dy()-watermark.Bounds().Dy()-10,
	)

	// 叠加水印 (0.9透明度)
	outputPath := s.TmpDir + "result.jpg"
	dst := imaging.Overlay(img, watermark, pos, 0.9)
	err = imaging.Save(dst, outputPath)
	if err != nil {
		return err
	}

	return nil
}

func (s *Storage) GetFileType(data []byte) string {
	kind, _ := filetype.Match(data)
	return kind.MIME.Value
}

func (s *Storage) Delete(ctx context.Context, key string) error {
	errAws := s.AwsDelete(ctx, key)
	errMinio := s.MinioDelete(ctx, key)

	if s.Provider == "aws" {
		return errAws
	}

	return errMinio
}
