package main

import (
	"context"
	"fmt"
	"sort"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

func getUniqueS3ObjectTags(ctx context.Context, flags Flags) ([]string, error) {
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(flags.Region))
	if err != nil {
		return nil, err
	}
	svc := s3.NewFromConfig(cfg)

	// List objects in the bucket
	result, err := svc.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
		Bucket: aws.String(flags.Bucket),
	})
	if err != nil {
		return nil, err
	}

	var wg sync.WaitGroup
	wg.Add(len(result.Contents))

	unique := make(map[string]bool)
	uniqueTags := []string{}
	mutex := sync.Mutex{}

	// Get tags for each object
	for _, obj := range result.Contents {
		go func(obj types.Object) {
			defer wg.Done()

			tagResult, err := svc.GetObjectTagging(ctx, &s3.GetObjectTaggingInput{
				Bucket: aws.String(flags.Bucket),
				Key:    obj.Key,
			})
			if err != nil {
				fmt.Printf("Error retrieving tags for object %s: %v\n", *obj.Key, err)
				return
			}

			for _, tag := range tagResult.TagSet {
				tagValue := *tag.Value
				mutex.Lock()
				if !unique[tagValue] {
					unique[tagValue] = true
					uniqueTags = append(uniqueTags, tagValue)
				}
				mutex.Unlock()
			}
		}(obj)
	}
	wg.Wait()

	return uniqueTags, nil
}

func sortTags(versions []string) []string {
	sortedVersions := make([]string, len(versions))
	copy(sortedVersions, versions)

	sort.SliceStable(sortedVersions, func(i, j int) bool {
		return compareVersions(sortedVersions[i], sortedVersions[j]) < 0
	})

	return sortedVersions
}

func getTagsToDelete(flags Flags, tags []string) []string {
	if len(tags) <= flags.NumLatestTagsToKeep {
		fmt.Println("Objects with these tags will remain:", tags)
		fmt.Println("No objects will be deleted.")
		return []string{}
	}
	tagsToKeep := tags[len(tags)-flags.NumLatestTagsToKeep:]
	tagsToDelete := tags[:len(tags)-flags.NumLatestTagsToKeep]

	fmt.Println("Objects with these tags will remain:", tagsToKeep)
	fmt.Println("Objects with these tags will be deleted:", tagsToDelete)

	return tagsToDelete
}
