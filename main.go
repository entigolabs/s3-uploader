package main

import (
	"context"
	"fmt"
)

func main() {
	ctx := context.Background()
	flags := Flags{}

	err := flags.getValues()
	if err != nil {
		fmt.Println("Error getting flag values:", err)
		return
	}

	err = validateAWSCredentials(ctx)
	if err != nil {
		fmt.Println("Error validating AWS credentials:", err)
		return
	}

	err = uploadFilesToS3(ctx, flags)
	if err != nil {
		fmt.Println("Error uploading files to S3:", err)
		return
	}

	uniqueTags, err := getUniqueS3ObjectTags(ctx, flags)
	if err != nil {
		fmt.Println("Error getting S3 object tags:", err)
		return
	}

	sortedUniqueTags := sortTags(uniqueTags)

	tagsToDelete := getTagsToDelete(flags, sortedUniqueTags)

	err = deleteObjectsWithTags(ctx, flags, tagsToDelete)
	if err != nil {
		fmt.Println("Error deleting objects with tags:", err)
		return
	}
}
