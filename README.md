# Empty Bucket Finalizer Service (Go)

This Go service monitors AWS S3 buckets managed by ACK (AWS Controllers for Kubernetes) for deletion events. If a bucket has a finalizer, the service empties the bucket before allowing deletion.

## Features
- Watches Kubernetes for `s3.services.k8s.aws/Bucket` resources with a finalizer and `deletionTimestamp`
- Empties all objects from the S3 bucket (no versioning)
- Removes the finalizer from the K8s resource after emptying

## Requirements
- Go 1.21+
- AWS credentials with S3 permissions (provided via Kubernetes service account or environment variables)
- Kubernetes cluster with ACK S3 controller installed

## Setup
1. Build the service:
   ```sh
   go build -o empty-bucket-finalizer main.go
   ```
2. Deploy to Kubernetes (see your previous YAML for deployment/service account setup).

## Usage
Run the service:
```sh
./empty-bucket-finalizer
```

## How It Works
- The service watches for `s3.services.k8s.aws/Bucket` resources with a finalizer and `deletionTimestamp`.
- When such a resource is found, it empties the S3 bucket and removes the finalizer from the K8s resource.
- This allows ACK to proceed with bucket deletion.

## Notes
- The service uses the Kubernetes dynamic client and AWS SDK for Go.
- Make sure your AWS credentials have permissions for S3 List and DeleteObject actions.
