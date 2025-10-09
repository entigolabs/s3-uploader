package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

func uploadFilesToS3(flags Flags) error {
	sess := session.Must(session.NewSession(&aws.Config{
		Region: aws.String(flags.Region),
	}))
	svc := s3.New(sess)

	var wg sync.WaitGroup
	concurrencyChan := make(chan struct{}, concurrentUploads)

	err := filepath.Walk(flags.SourceDirectory, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			concurrencyChan <- struct{}{}
			wg.Add(1)
			go func(sourceDir string, path string, bucket string, tag string, svc *s3.S3, wg *sync.WaitGroup) {
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
				defer file.Close()

				contentType, err := getContentType(file)
				if err != nil {
					fmt.Println("Error getting content type:", err, "File:", path)
					return
				}
				file.Seek(0, 0)

				object := &s3.PutObjectInput{
					Bucket:       aws.String(bucket),
					Key:          aws.String(s3Path),
					Body:         file,
					ContentType:  aws.String(contentType),
					CacheControl: aws.String(defaultCacheControl),
					Tagging:      aws.String(tag),
				}

				fileIsIndexHTML, _ := filepath.Match("/index.html", s3Path)
				if fileIsIndexHTML {
					object.CacheControl = aws.String(indexHTMLCacheControl)
				}

				fmt.Println("Uploading:", s3Path, "ContentType:", *object.ContentType, "Tag:", *object.Tagging)

				_, err = svc.PutObject(object)
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
