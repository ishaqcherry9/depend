package oss

import (
	"context"
	"fmt"
	"log"
	"testing"
)

// go test -v -test.run TestOss
func TestOss(t *testing.T) {
	ctx := context.Background()
	storage, err := NewStorage(ctx, "http://192.168.100.165:8788/api/v1/oss/getConfig")
	if err != nil {
		log.Fatal(err)
	}

	objects, err := storage.List(ctx)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(objects)

	upload := UploadOssInfo{
		Key:         "20251013/abcdef.mp4",
		FileName:    "/Users/robert/Downloads/ql_1758374133263_mp4.mp4",
		Acl:         storage.Aws.AwsAcl,
		ContentType: "video/mp4",
		//ContentType: "image/jpeg",
	}

	downloadURL, err := storage.Upload(ctx, upload)
	if err != nil {
		log.Fatal(err)
		return
	}

	fmt.Println("result: ", downloadURL)

	body, err := storage.Get(ctx, "20251013/abcdef.mp4")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("get len:", len(body))

	if err := storage.Delete(ctx, "20251013/abcdef.mp4"); err != nil {
		log.Fatal(err)
	}
}
