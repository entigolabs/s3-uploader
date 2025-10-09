package main

import (
	"flag"
	"fmt"
)

type Flags struct {
	NumLatestTagsToKeep int
	SourceDirectory     string
	TargetDirectory     string
	Bucket              string
	Region              string
	Tag                 string
}

func (c *Flags) getValues() error {
	flag.IntVar(&c.NumLatestTagsToKeep, "num-latest-tags-to-keep", 0, "Number of latest tags to keep")
	flag.StringVar(&c.SourceDirectory, "source-directory", "", "Source directory")
	flag.StringVar(&c.TargetDirectory, "target-directory", "", "Target directory")
	flag.StringVar(&c.Bucket, "bucket", "", "AWS bucket name")
	flag.StringVar(&c.Region, "region", "", "AWS region")
	flag.StringVar(&c.Tag, "tag", "", "Tag")
	flag.Parse()

	if c.NumLatestTagsToKeep == 0 || c.SourceDirectory == "" || c.TargetDirectory == "" || c.Bucket == "" || c.Region == "" || c.Tag == "" {
		return fmt.Errorf("all flags must be set")
	}

	fmt.Println("Number of latest tags to keep:", c.NumLatestTagsToKeep)
	fmt.Println("Source directory:", c.SourceDirectory)
	fmt.Println("Target directory:", c.TargetDirectory)
	fmt.Println("AWS bucket name:", c.Bucket)
	fmt.Println("AWS region:", c.Region)
	fmt.Println("Tag:", c.Tag)

	return nil
}
