# ACK S3 Empty Bucket Controller

This service monitors AWS S3 buckets managed by ACK (AWS Controllers for Kubernetes) for deletion events. If a bucket has the annotation `s3.services.k8s.aws/empty-on-delete: "true"`, the service empties the bucket before deletion.

## Usage

The service uses in-cluster config by default. For local development, set the `KUBECONFIG` environment variable or ensure `~/.kube/config` is present.

AWS credentials are loaded using the default AWS SDK provider chain (env vars, IAM roles, etc).

