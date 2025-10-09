# s3-uploader

s3-uploader is a CLI application written in Go that will upload files to an S3 bucket and add appropriate tags and metadata to the uploaded files. These tags will be used to identify the files that should remain in the bucket and the files that should be deleted.

It will set following metadata for the uploaded files: `max-age=2592000,public` <br>
It will set following metadata for the index.html file: `nocache`

## Example usage

Set the environment variables:

```
export AWS_ACCESS_KEY_ID=<my-aws-access-key-id>
export AWS_SECRET_ACCESS_KEY=<my-aws-secret-access-key>
```

Build application with `go build`

Available flags:

- `--bucket`: S3 bucket name
- `--region`: AWS region
- `--source-directory`: Source directory
- `--target-directory`: Target directory
- `--num-latest-tags-to-keep`: Number of latest tags to keep. This is used to identify the files that should remain in the bucket and the files that should be deleted.
- `--tag`: Tag to add to the uploaded files

```
./s3-uploader --bucket mybucket --region eu-west-1 --source-directory source/ --target-directory target/ --num-latest-tags-to-keep 3 --tag version=1.0.0
```
