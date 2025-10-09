package main

import (
	"fmt"
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

func deleteObjectsWithTags(flags Flags, tagsToDelete []string) error {
	sess := session.Must(session.NewSession(&aws.Config{
		Region: aws.String(flags.Region),
	}))
	svc := s3.New(sess)

	input := &s3.ListObjectsInput{
		Bucket: aws.String(flags.Bucket),
	}

	err := svc.ListObjectsPages(input,
		func(page *s3.ListObjectsOutput, lastPage bool) bool {
			var wg sync.WaitGroup
			concurrencyChan := make(chan struct{}, concurrentDeletions)

			for _, obj := range page.Contents {
				wg.Add(1)
				concurrencyChan <- struct{}{}
				go func(obj *s3.Object) {
					defer func() {
						<-concurrencyChan
						wg.Done()
					}()

					// For each object, get its tags and check if any matches the tags to delete
					input := &s3.GetObjectTaggingInput{
						Bucket: aws.String(flags.Bucket),
						Key:    obj.Key,
					}

					tagResult, err := svc.GetObjectTagging(input)
					if err != nil {
						fmt.Printf("Error retrieving tags for object %s: %v\n", *obj.Key, err)
						return
					}

					for _, tag := range tagResult.TagSet {
						if contains(tagsToDelete, *tag.Value) {
							// If the object has a tag to delete, delete the object
							deleteInput := &s3.DeleteObjectInput{
								Bucket: aws.String(flags.Bucket),
								Key:    obj.Key,
							}

							_, err := svc.DeleteObject(deleteInput)
							if err != nil {
								fmt.Printf("Error deleting object %s: %v\n", *obj.Key, err)
								return
							}
							fmt.Printf("Deleted object: %s, Tag: %s\n", *obj.Key, *tag.Value)
						}
					}
				}(obj)
			}
			wg.Wait()
			return true
		})
	if err != nil {
		return err
	}
	return nil
}
