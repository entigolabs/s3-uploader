package main

import (
	"fmt"
)

const (
	concurrentUploads     = 500
	concurrentDeletions   = 500
	defaultCacheControl   = "max-age=31536000,public"
	indexHTMLCacheControl = "nocache"
)

func main() {
	flags := Flags{}

	err := flags.getValues()
	if err != nil {
		fmt.Println("Error getting flag values:", err)
		return
	}

	err = validateAWSCredentials()
	if err != nil {
		fmt.Println("Error validating AWS credentials:", err)
		return
	}

	err = uploadFilesToS3(flags)
	if err != nil {
		fmt.Println("Error uploading files to S3:", err)
		return
	}

	uniqueTags, err := getUniqueS3ObjectTags(flags)
	if err != nil {
		fmt.Println("Error getting S3 object tags:", err)
		return
	}

	sortedUniqueTags := sortTags(uniqueTags)

	tagsToDelete := getTagsToDelete(flags, sortedUniqueTags)

	err = deleteObjectsWithTags(flags, tagsToDelete)
	if err != nil {
		fmt.Println("Error deleting objects with tags:", err)
		return
	}
}
