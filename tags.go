package main

import (
	"fmt"
	"sort"
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

func getUniqueS3ObjectTags(flags Flags) ([]string, error) {
	sess := session.Must(session.NewSession(&aws.Config{
		Region: aws.String(flags.Region),
	}))
	svc := s3.New(sess)

	// List objects in the assets folder
	input := &s3.ListObjectsInput{
		Bucket: aws.String(flags.Bucket),
	}

	result, err := svc.ListObjects(input)
	if err != nil {
		return nil, err
	}

	var wg sync.WaitGroup
	wg.Add(len(result.Contents))

	unique := make(map[string]bool)
	uniqueTags := []string{}
	mutex := sync.Mutex{}

	// Get tags for each object in assets folder
	for _, obj := range result.Contents {
		go func(obj *s3.Object) {
			defer wg.Done()

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
