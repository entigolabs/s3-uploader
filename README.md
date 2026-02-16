# s3-uploader

CLI tool for uploading files to S3 with version tagging and automatic cleanup of old versions.
Tags are sorted as semver and oldest versions are deleted based on `--num-latest-tags-to-keep`.

## Flags

### Required

| Flag                        | Description                                  |
| --------------------------- | -------------------------------------------- |
| `--bucket`                  | S3 bucket name                               |
| `--region`                  | AWS region                                   |
| `--source-directory`        | Local source directory                       |
| `--num-latest-tags-to-keep` | Number of latest tags to keep                |
| `--tag`                     | Tag for uploaded files (format: `key=value`) |

### Optional

| Flag                     | Default                   | Description                         |
| ------------------------ | ------------------------- | ----------------------------------- |
| `--target-directory`     | `""` (bucket root)        | S3 target directory                 |
| `--concurrent-uploads`   | `500`                     | Number of concurrent uploads        |
| `--concurrent-deletions` | `500`                     | Number of concurrent deletions      |
| `--cache-control`        | `max-age=31536000,public` | Cache-Control header for files      |
| `--index-cache-control`  | `no-cache`                | Cache-Control header for index.html |

## Docker Images

| Image                                    | Description                      |
| ---------------------------------------- | -------------------------------- |
| `ghcr.io/entigolabs/s3-uploader`         | Base image with s3-uploader only |
| `ghcr.io/entigolabs/s3-uploader-aws-cli` | Includes AWS CLI                 |

## Quick Start (Docker)

```
export AWS_ACCESS_KEY_ID=<my-aws-access-key-id>
export AWS_SECRET_ACCESS_KEY=<my-aws-secret-access-key>

docker run \
  -e AWS_ACCESS_KEY_ID \
  -e AWS_SECRET_ACCESS_KEY \
  -v $(pwd)/source:/source \
  ghcr.io/entigolabs/s3-uploader:latest \
  --bucket mybucket \
  --region eu-west-1 \
  --source-directory /source \
  --target-directory target/ \
  --num-latest-tags-to-keep 3 \
  --tag version=1.0.0
```

## CI/CD Usage (with AWS CLI)

```yaml
# Bitbucket Pipelines example
- step:
    name: Deploy
    image: ghcr.io/entigolabs/s3-uploader-aws-cli:latest
    script:
      - s3-uploader --bucket mybucket --region eu-north-1 --source-directory source --num-latest-tags-to-keep 3 --tag version=1.0.0
      - aws cloudfront create-invalidation --distribution-id XXXXX --paths '/*'
```
