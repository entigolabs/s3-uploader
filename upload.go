package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func uploadFilesToS3(ctx context.Context, flags Flags) error {
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(flags.Region))
	if err != nil {
		return err
	}
	svc := s3.NewFromConfig(cfg)

	var wg sync.WaitGroup
	concurrencyChan := make(chan struct{}, flags.ConcurrentUploads)

	err = filepath.Walk(flags.SourceDirectory, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			concurrencyChan <- struct{}{}
			wg.Add(1)
			go func(sourceDir string, path string, bucket string, tag string, svc *s3.Client, wg *sync.WaitGroup) {
				defer func() {
					<-concurrencyChan
					wg.Done()
				}()

				relativePath, _ := filepath.Rel(sourceDir, path)
				s3Path := filepath.ToSlash(filepath.Join(flags.TargetDirectory, relativePath))

				file, err := os.Open(path)
				if err != nil {
					fmt.Println("Error opening file:", err)
					return
				}
				defer func() { _ = file.Close() }()

				contentType, err := getContentType(file)
				if err != nil {
					fmt.Println("Error getting content type:", err, "File:", path)
					return
				}
				if _, err := file.Seek(0, 0); err != nil {
					fmt.Println("Error seeking file:", err, "File:", path)
					return
				}

				cacheControl := flags.DefaultCacheControl
				if filepath.Base(s3Path) == "index.html" {
					cacheControl = flags.IndexHTMLCacheControl
				}

				fmt.Println("Uploading:", s3Path, "ContentType:", contentType, "Tag:", tag)

				_, err = svc.PutObject(ctx, &s3.PutObjectInput{
					Bucket:       aws.String(bucket),
					Key:          aws.String(s3Path),
					Body:         file,
					ContentType:  aws.String(contentType),
					CacheControl: aws.String(cacheControl),
					Tagging:      aws.String(tag),
				})
				if err != nil {
					fmt.Println("Error uploading file:", err, "File:", path)
					return
				}
			}(flags.SourceDirectory, path, flags.Bucket, flags.Tag, svc, &wg)
		}
		return nil
	})
	wg.Wait()
	if err != nil {
		return err
	}
	return nil
}
