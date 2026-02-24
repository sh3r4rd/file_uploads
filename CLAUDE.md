# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Serverless file upload service on AWS using presigned URLs for secure PDF uploads. Two-step upload flow: client requests a presigned URL via API Gateway/Lambda, then PUTs the file directly to S3. An S3 event triggers a second Lambda to validate and confirm the upload.

## Build & Development Commands

```bash
make build          # Cross-compile Go lambdas for Amazon Linux 2023 ARM64
make test           # Run all Go tests (cd lambdas && go test ./...)
make lint           # Run go vet on lambdas
make fmt            # Format Go code + Terraform files
make init           # Initialize Terraform (cd infra && terraform init)
make validate       # Validate Terraform config
make deploy         # Build + deploy via Terraform
make clean          # Remove compiled binaries (lambdas/bin/)
```

Run a single test:
```bash
cd lambdas && go test ./internal/validation/ -run TestValidateName
```

## Architecture

**Request flow:**
1. `POST /uploads` → API Gateway (API key auth) → `RequestUploadHandler` Lambda → validates input, writes PENDING record to DynamoDB, returns presigned S3 PUT URL
2. Client PUTs PDF directly to S3 using presigned URL
3. S3 event notification → `ConfirmUploadHandler` Lambda → validates PDF (magic bytes, size), updates DynamoDB status to UPLOADED or REJECTED

**Two Lambda functions** (Go, `provided.al2023` runtime, ARM64):
- `lambdas/cmd/request-upload/` — API Gateway trigger, generates presigned URLs
- `lambdas/cmd/confirm-upload/` — S3 event trigger, validates uploaded files

**Shared Go packages** in `lambdas/internal/`:
- `handler/` — core handler logic for each Lambda
- `validation/` — file validation rules (type, size, PDF magic bytes)
- `storage/` — S3 and DynamoDB operations
- `port/` — interfaces for external dependencies (enables mocking in tests)
- `model/` — request/response structs, DynamoDB item types

**Infrastructure** (`infra/`) — Terraform HCL, split by resource type: `s3.tf`, `dynamodb.tf`, `sqs.tf`, `iam.tf`, `lambda.tf`, `apigateway.tf`. Each Lambda gets its own IAM role with least-privilege permissions scoped to specific resource ARNs.

## Key Technical Details

- Go module: `github.com/sh3r4rd/file_uploads` (in `lambdas/`)
- Uses `aws-sdk-go-v2` (not v1) and `aws-lambda-go`
- S3 key structure: `uploads/{userId}/{fileId}.pdf`
- DynamoDB table: `FileUploads` with `fileId` (PK) and GSI on `userId-createdAt`
- Presigned URLs have 5-minute TTL with content-type and content-length conditions
- File constraints: PDF only (`application/pdf`), max 1 MB (1,048,576 bytes)
- Terraform variables configured via `infra/terraform.tfvars` (copy from `terraform.tfvars.example`)

## Prerequisites

- Go 1.25+ (go.mod specifies 1.25.1; latest stable is 1.26)
- Terraform >= 1.5 (latest stable is 1.14.x)
- AWS CLI configured with appropriate credentials
