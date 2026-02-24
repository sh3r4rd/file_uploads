# Serverless File Upload Service

A serverless file upload system built on AWS using presigned URLs for secure, scalable PDF uploads.

## Architecture

```
Client → API Gateway → Lambda (RequestUpload) → Presigned URL
Client → S3 (direct PUT via presigned URL)
S3 Event → Lambda (ConfirmUpload) → DynamoDB status update
```

**Services:** API Gateway, Lambda (Go), S3, DynamoDB, SQS

## Directory Structure

```
.
├── infra/                          # Terraform infrastructure
│   ├── providers.tf
│   ├── variables.tf
│   ├── main.tf
│   ├── outputs.tf
│   ├── s3.tf
│   ├── dynamodb.tf
│   ├── sqs.tf
│   ├── iam.tf
│   ├── lambda.tf
│   ├── apigateway.tf
│   └── terraform.tfvars.example
├── lambdas/                        # Go Lambda functions
│   ├── cmd/
│   │   ├── request-upload/         # API Gateway handler
│   │   └── confirm-upload/         # S3 event handler
│   └── internal/
│       ├── handler/
│       ├── validation/
│       ├── storage/
│       ├── port/
│       └── model/
├── scripts/
│   ├── build.sh
│   └── deploy.sh
└── Makefile
```

## Prerequisites

- Go 1.22+
- Terraform >= 1.5
- AWS CLI configured with appropriate credentials

## Quick Start

```bash
# Initialize Terraform
make init

# Format code
make fmt

# Run tests
make test

# Lint
make lint

# Build Lambda binaries
make build

# Deploy
make deploy
```

## Configuration

Copy the example tfvars file and edit as needed:

```bash
cp infra/terraform.tfvars.example infra/terraform.tfvars
```

| Variable | Default | Description |
|----------|---------|-------------|
| `aws_region` | `us-east-1` | AWS region for all resources |
| `project_name` | `file-uploads` | Project name for resource naming |
| `environment` | `dev` | Deployment environment |
