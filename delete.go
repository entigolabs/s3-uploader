package main

import (
	"context"
	"fmt"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

func deleteObjectsWithTags(ctx context.Context, flags Flags, tagsToDelete []string) error {
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(flags.Region))
	if err != nil {
		return err
	}
	svc := s3.NewFromConfig(cfg)

	paginator := s3.NewListObjectsV2Paginator(svc, &s3.ListObjectsV2Input{
		Bucket: aws.String(flags.Bucket),
	})

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return err
		}

		var wg sync.WaitGroup
		concurrencyChan := make(chan struct{}, flags.ConcurrentDeletions)

		for _, obj := range page.Contents {
			wg.Add(1)
			concurrencyChan <- struct{}{}
			go func(obj types.Object) {
				defer func() {
					<-concurrencyChan
					wg.Done()
				}()

				// For each object, get its tags and check if any matches the tags to delete
				tagResult, err := svc.GetObjectTagging(ctx, &s3.GetObjectTaggingInput{
					Bucket: aws.String(flags.Bucket),
					Key:    obj.Key,
				})
				if err != nil {
					fmt.Printf("Error retrieving tags for object %s: %v\n", *obj.Key, err)
					return
				}

				for _, tag := range tagResult.TagSet {
					if contains(tagsToDelete, *tag.Value) {
						// If the object has a tag to delete, delete the object
						_, err := svc.DeleteObject(ctx, &s3.DeleteObjectInput{
							Bucket: aws.String(flags.Bucket),
							Key:    obj.Key,
						})
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
	}
	return nil
}
